package sessions

import "time"

type Status string

const (
	StatusRunning       Status = "running"
	StatusNeedsInput    Status = "needs_input"
	StatusNeedsApproval Status = "needs_approval"
	StatusError         Status = "error"
	StatusStopped       Status = "stopped"
	StatusUnknown       Status = "unknown"
)

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RepoPath  string    `json:"repoPath"`
	TmuxName  string    `json:"tmuxName"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
