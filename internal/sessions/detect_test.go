package sessions

import "testing"

func TestDetectStatus(t *testing.T) {
	tests := []struct {
		name   string
		exists bool
		output string
		want   Status
	}{
		{name: "stopped", exists: false, want: StatusStopped},
		{name: "approval", exists: true, output: "please approve this command", want: StatusNeedsApproval},
		{name: "continue", exists: true, output: "continue?", want: StatusNeedsInput},
		{name: "yes no", exists: true, output: "Proceed? y/N", want: StatusNeedsInput},
		{name: "permission", exists: true, output: "permission required", want: StatusNeedsInput},
		{name: "error", exists: true, output: "error: bad state", want: StatusError},
		{name: "failed", exists: true, output: "command failed", want: StatusError},
		{name: "running", exists: true, output: "working", want: StatusRunning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectStatus(tt.exists, tt.output); got != tt.want {
				t.Fatalf("DetectStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
