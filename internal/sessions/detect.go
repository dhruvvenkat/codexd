package sessions

import "strings"

func DetectStatus(tmuxExists bool, output string) Status {
	if !tmuxExists {
		return StatusStopped
	}

	text := strings.ToLower(output)
	switch {
	case strings.Contains(text, "approve"):
		return StatusNeedsApproval
	case strings.Contains(text, "continue?"):
		return StatusNeedsInput
	case strings.Contains(text, "y/n"):
		return StatusNeedsInput
	case strings.Contains(text, "permission"):
		return StatusNeedsInput
	case strings.Contains(text, "error"):
		return StatusError
	case strings.Contains(text, "failed"):
		return StatusError
	default:
		return StatusRunning
	}
}
