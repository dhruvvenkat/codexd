package sessions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTmuxSendUsesLiteralTextThenCarriageReturn(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "args.log")
	binPath := filepath.Join(dir, "tmux-fake")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + logPath + "\n"
	if err := os.WriteFile(binPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake tmux: %v", err)
	}

	client := TmuxClient{Bin: binPath}
	if err := client.Send("codexd-demo", "fix tests -- carefully"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	got := strings.Split(strings.TrimSpace(string(b)), "\n")
	want := []string{
		"send-keys -t codexd-demo -l fix tests -- carefully",
		"send-keys -t codexd-demo C-m",
	}
	if len(got) != len(want) {
		t.Fatalf("logged commands = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("command %d = %q, want %q", i, got[i], want[i])
		}
	}
}
