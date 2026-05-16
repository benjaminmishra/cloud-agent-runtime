# Runtime Architecture

CLI -> Backend API/WebSocket -> Runtime abstraction -> Docker Sandbox -> Agent Process (aider-chat)

## Components
- `cli/`: local-native command UX.
- `backend/`: session API, websocket stream, lifecycle handlers.
- `runtime/`: sandbox interface + docker implementation.
- `sandbox/`: immutable runtime image.
- `shared/`: transport/session contracts.
