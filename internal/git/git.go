package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Client struct {
	Bin string
}

func (c Client) bin() string {
	if c.Bin == "" {
		return "git"
	}
	return c.Bin
}

func (c Client) Status(repoPath string) (string, error) {
	return c.run(repoPath, "status", "--short")
}

func (c Client) Diff(repoPath string) (string, error) {
	return c.run(repoPath, "diff")
}

func (c Client) run(repoPath string, args ...string) (string, error) {
	all := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command(c.bin(), all...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return string(out), nil
}
