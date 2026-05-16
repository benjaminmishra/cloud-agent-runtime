# Rune: Remote Agent Runtime PoC

## 1) Product vision
Rune is an AI-agent-agnostic remote execution runtime that keeps unsafe AI-generated code off developer machines by running agent workloads in ephemeral cloud-like Docker sandboxes.

## 2) Architecture diagram
```text
CLI (rune)
  -> Backend API + WebSocket
    -> Sandbox Runtime Interface
      -> Docker Container Sandbox
        -> Agent Process (aider-chat)
```

## 3) Local setup
- Go 1.22+
- Docker Engine
- Docker Compose

## 4) Quick start
```bash
make build
docker-compose up --build
./bin/rune run https://github.com/foo/bar --backend http://localhost:8080
```

## 5) Example CLI usage
```bash
rune run https://github.com/example/repo
```
This triggers remote sandbox start, repository clone, agent startup, websocket terminal streaming, Ctrl+C stop, and cleanup.

## 6) Security model
- Ephemeral, isolated Docker containers.
- No host source mount into sandbox.
- Auto-remove containers on stop.
- Resource/policy hooks in runtime layer for CPU, memory, timeout, network, and policy profile.

## 7) Runtime architecture
See `docs/architecture.md` and `runtime/sandbox.go` abstraction for future runtime backends (Firecracker/K8s/VMs).

## 8) Future roadmap
- Firecracker microVM runtime
- enterprise governance & policy engine
- audit logging/event pipeline
- browser isolation
- remote IDE integration
- runtime policy DSL
- agent runtime protocol
- multi-tenant execution
- observability + OpenTelemetry
- event sourcing for lifecycle state
