package runtime

import "context"

type Sandbox interface {
	Start(ctx context.Context) error
	Exec(ctx context.Context, cmd string) error
	Stream(ctx context.Context) (<-chan []byte, error)
	Stop(ctx context.Context) error
	ContainerID() string
}

type ResourceLimits struct {
	CPUQuota       int64
	MemoryBytes    int64
	ExecutionLimit int64
	NetworkPolicy  string
	PolicyProfile  string
}
