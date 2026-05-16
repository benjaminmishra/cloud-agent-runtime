package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud-agent-runtime/runtime"
	"cloud-agent-runtime/shared"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

type SessionRuntime struct {
	Session shared.Session
	Sandbox runtime.Sandbox
}

type SessionManager struct {
	docker *client.Client
	items  map[string]*SessionRuntime
	mu     sync.RWMutex
}

func NewSessionManager() (*SessionManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &SessionManager{docker: cli, items: make(map[string]*SessionRuntime)}, nil
}

func (m *SessionManager) Create(ctx context.Context, repoURL string) (*SessionRuntime, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	s := shared.Session{ID: id, RepoURL: repoURL, Status: shared.SessionStatusStarting, CreatedAt: now, UpdatedAt: now}
	sb := runtime.NewDockerSandbox(m.docker, "rune-sandbox:latest", runtime.ResourceLimits{CPUQuota: 50000, MemoryBytes: 2 * 1024 * 1024 * 1024, ExecutionLimit: 3600, NetworkPolicy: "default", PolicyProfile: "poc"})
	entry := &SessionRuntime{Session: s, Sandbox: sb}
	m.mu.Lock()
	m.items[id] = entry
	m.mu.Unlock()

	if err := sb.Start(ctx); err != nil {
		m.markFailed(id, err)
		return nil, err
	}
	entry.Session.ContainerID = sb.ContainerID()
	if err := sb.Exec(ctx, fmt.Sprintf("cd /workspace && rm -rf repo && git clone %q repo", repoURL)); err != nil {
		m.markFailed(id, err)
		return nil, err
	}
	entry.Session.Status = shared.SessionStatusRunning
	entry.Session.UpdatedAt = time.Now().UTC()
	return entry, nil
}

func (m *SessionManager) Get(id string) (*SessionRuntime, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.items[id]
	return s, ok
}

func (m *SessionManager) Stop(ctx context.Context, id string) error {
	m.mu.RLock()
	s, ok := m.items[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("not found")
	}
	if err := s.Sandbox.Stop(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	s.Session.Status = shared.SessionStatusStopped
	s.Session.UpdatedAt = time.Now().UTC()
	m.mu.Unlock()
	return nil
}

func (m *SessionManager) markFailed(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.items[id]; ok {
		s.Session.Status = shared.SessionStatusFailed
		s.Session.Error = err.Error()
		s.Session.UpdatedAt = time.Now().UTC()
	}
}
