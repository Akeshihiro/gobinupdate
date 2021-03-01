package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if !isGoInstalled() {
		fmt.Println("no Go compiler installed")
		return
	}

	installedGoTools, err := listInstalledGoTools()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, t := range installedGoTools {
		src, err := determineInstallationSource(t)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = updateGoTool(src)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func isGoInstalled() bool {
	cmd := exec.Command("go", "version")
	err := cmd.Run()

	return err == nil
}

func getGoBinPath() (string, error) {
	gopath := getGoEnv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("env variable 'GOPATH' not set or empty")
	}

	gobin := gopath + "/bin"
	_, err := os.Stat(gobin)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("path '%v' does not exist", gobin)
	}

	return gobin, nil
}

func getGoEnv(s string) string {
	cmd := exec.Command("go", "env", s)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

func listInstalledGoTools() ([]string, error) {
	gobinpath, err := getGoBinPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(gobinpath)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		result = append(result, filepath.Join(gobinpath, e.Name()))
	}

	return result, nil
}

func determineInstallationSource(toolpath string) (string, error) {
	cmd := exec.Command("go", "tool", "objdump", "-s", "main.main", toolpath)
	tmpOutput, err := cmd.Output()
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(string(tmpOutput))
	line := strings.TrimSpace(strings.Split(output, "\n")[0])
	if !strings.Contains(line, "@") {
		return "", fmt.Errorf("'%v' was not installed by 'go install' command", toolpath)
	}

	gomodcachedirpath, err := getGoModCacheDirPath()
	if err != nil {
		return "", err
	}

	modsrc := strings.TrimSpace(strings.Fields(line)[2])
	modsrc = strings.TrimPrefix(modsrc, gomodcachedirpath+string(filepath.Separator))

	idx := strings.IndexRune(modsrc, '@')
	firstHalf := modsrc[:idx]
	secondHalf := modsrc[idx+1:]
	secondHalf = secondHalf[strings.IndexRune(secondHalf, filepath.Separator):]
	secondHalf = secondHalf[:strings.LastIndex(secondHalf, string(filepath.Separator))]

	modsrc = firstHalf + secondHalf
	modsrc = strings.ReplaceAll(modsrc, string(filepath.Separator), "/")

	return modsrc, nil
}

func getGoModCacheDirPath() (string, error) {
	gomodcachedirpath := getGoEnv("GOMODCACHE")
	if gomodcachedirpath == "" {
		return "", fmt.Errorf("env var 'GOMODCACHE' not set or empty")
	}

	return gomodcachedirpath, nil
}

func updateGoTool(src string) error {
	cmd := exec.Command("go", "install", src+"@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
