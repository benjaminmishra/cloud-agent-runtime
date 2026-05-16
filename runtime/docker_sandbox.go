package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type DockerSandbox struct {
	cli         *client.Client
	image       string
	workspace   string
	containerID string
	limits      ResourceLimits
	mu          sync.RWMutex
}

func NewDockerSandbox(cli *client.Client, imageName string, limits ResourceLimits) *DockerSandbox {
	if imageName == "" {
		imageName = "rune-sandbox:latest"
	}
	return &DockerSandbox{cli: cli, image: imageName, workspace: "/workspace", limits: limits}
}

func (d *DockerSandbox) Start(ctx context.Context) error {
	pull, err := d.cli.ImagePull(ctx, d.image, image.PullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pull)
		_ = pull.Close()
	}

	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:        d.image,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   d.workspace,
		Cmd:          []string{"bash", "-lc", "tail -f /dev/null"},
	}, &container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			Memory:   d.limits.MemoryBytes,
			CPUQuota: d.limits.CPUQuota,
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.containerID = resp.ID
	d.mu.Unlock()

	return d.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

func (d *DockerSandbox) Exec(ctx context.Context, cmd string) error {
	id := d.ContainerID()
	if id == "" {
		return fmt.Errorf("container not started")
	}
	execCfg := container.ExecOptions{Cmd: []string{"bash", "-lc", cmd}, AttachStdout: true, AttachStderr: true}
	hijack, err := d.cli.ContainerExecAttach(ctx, id, execCfg)
	if err != nil {
		return err
	}
	defer hijack.Close()
	_, err = io.Copy(io.Discard, hijack.Reader)
	return err
}

func (d *DockerSandbox) Stream(ctx context.Context) (<-chan []byte, error) {
	id := d.ContainerID()
	if id == "" {
		return nil, fmt.Errorf("container not started")
	}
	out := make(chan []byte, 64)
	execCfg := container.ExecOptions{Cmd: []string{"bash", "-lc", "cd /workspace && if [ -d repo ]; then cd repo; fi; aider --help || echo 'sandbox ready';"}, AttachStdout: true, AttachStderr: true, Tty: true}
	hijack, err := d.cli.ContainerExecAttach(ctx, id, execCfg)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(out)
		defer hijack.Close()
		scanner := bufio.NewScanner(hijack.Reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r\n")
			select {
			case out <- append([]byte(line), '\n'):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (d *DockerSandbox) Stop(ctx context.Context) error {
	id := d.ContainerID()
	if id == "" {
		return nil
	}
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func (d *DockerSandbox) ContainerID() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.containerID
}
