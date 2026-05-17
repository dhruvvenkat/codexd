package sessions

import (
	"fmt"
	"os/exec"
	"strings"
)

type Tmux interface {
	Start(tmuxName, repoPath, codexCommand string) error
	Exists(tmuxName string) bool
	Capture(tmuxName string, lines int) (string, error)
	Send(tmuxName, text string) error
	Kill(tmuxName string) error
}

type TmuxClient struct {
	Bin string
}

func (c TmuxClient) bin() string {
	if c.Bin == "" {
		return "tmux"
	}
	return c.Bin
}

func (c TmuxClient) Start(tmuxName, repoPath, codexCommand string) error {
	if codexCommand == "" {
		codexCommand = "codex"
	}
	return run(c.bin(), "new-session", "-d", "-s", tmuxName, "-c", repoPath, codexCommand)
}

func (c TmuxClient) Exists(tmuxName string) bool {
	err := exec.Command(c.bin(), "has-session", "-t", tmuxName).Run()
	return err == nil
}

func (c TmuxClient) Capture(tmuxName string, lines int) (string, error) {
	if lines <= 0 {
		lines = 200
	}
	cmd := exec.Command(c.bin(), "capture-pane", "-t", tmuxName, "-p", "-S", fmt.Sprintf("-%d", lines))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandError(c.bin(), err, out)
	}
	return string(out), nil
}

func (c TmuxClient) Send(tmuxName, text string) error {
	if err := run(c.bin(), sendLiteralArgs(tmuxName, text)...); err != nil {
		return err
	}
	return run(c.bin(), sendSubmitArgs(tmuxName)...)
}

func (c TmuxClient) Kill(tmuxName string) error {
	return run(c.bin(), "kill-session", "-t", tmuxName)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return commandError(name, err, out)
	}
	return nil
}

func commandError(name string, err error, output []byte) error {
	msg := strings.TrimSpace(string(output))
	if msg == "" {
		return fmt.Errorf("%s: %w", name, err)
	}
	return fmt.Errorf("%s: %s", name, msg)
}

func sendLiteralArgs(tmuxName, text string) []string {
	return []string{"send-keys", "-t", tmuxName, "-l", text}
}

func sendSubmitArgs(tmuxName string) []string {
	return []string{"send-keys", "-t", tmuxName, "C-m"}
}
