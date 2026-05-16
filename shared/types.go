package shared

import "time"

type SessionStatus string

const (
	SessionStatusStarting SessionStatus = "starting"
	SessionStatusRunning  SessionStatus = "running"
	SessionStatusStopped  SessionStatus = "stopped"
	SessionStatusFailed   SessionStatus = "failed"
)

type Session struct {
	ID          string        `json:"id"`
	RepoURL     string        `json:"repoUrl"`
	Status      SessionStatus `json:"status"`
	ContainerID string        `json:"containerId,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	Error       string        `json:"error,omitempty"`
}

type CreateSessionRequest struct {
	RepoURL string `json:"repoUrl" binding:"required"`
}

type CreateSessionResponse struct {
	Session Session `json:"session"`
}
