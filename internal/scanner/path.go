package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExpandHome expands the ~ in a path to the user's home directory
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	expandedPath := ExpandHome(path)
	_, err := os.Stat(expandedPath)
	return err == nil
}

// FindExecutable finds an executable in the system PATH
func FindExecutable(name string) (string, error) {
	return exec.LookPath(name)
}

// GetExecutableVersion runs a command to get version information
func GetExecutableVersion(executable string, args ...string) (string, error) {
	cmd := exec.Command(executable, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ResolveSymlink resolves a symlink to its target
func ResolveSymlink(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

// GetEnvVar gets an environment variable value
func GetEnvVar(name string) string {
	return os.Getenv(name)
}
