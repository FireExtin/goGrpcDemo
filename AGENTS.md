# Repository Guidelines

## Project Structure

- `*.go` in the repo root are small, standalone demos (each file has its own `main()`).
- `handler/` contains RPC service implementations (e.g., `handler.HelloService`).
- `serverStub/` wraps registration helpers for `net/rpc` (shared by servers).
- `proto/` contains protobuf definitions (currently `proto/helloworld.proto`).
- Python demo clients live in `py*.py`; a local Python environment exists in `venv/` (treat as dev-only).

## Build, Test, and Development Commands

This repository is a set of runnable examples rather than a single buildable binary.

- Run TCP `net/rpc` server: `go run server.go` (listens on `:1234`)
- Run JSON-RPC server over TCP: `go run jsonRpcServer.go` (listens on `:1234`)
- Run JSON-RPC server over HTTP: `go run httpServer.go` (listens on `:8082`, endpoint `/jsonRpc`)
- Run Go clients: `go run tcpClient.go` or `go run jsonRpcClient.go`
- Run gRPC server/client: `go run ./cmd/grpcserver` and `go run ./cmd/grpcclient -name=test`
- Run Python clients: `python3 pyJsonRpcClient.py` (TCP) or `python3 pyhttpRpcClient.py` (HTTP)
- Static checks: `go fmt ./...` and `go vet ./handler ./serverStub` (or `go vet ./...` after moving root programs into `cmd/`)

## Coding Style & Naming Conventions

- Go: format with `gofmt` (tabs for indentation, standard Go style). Keep exported names `PascalCase`, unexported `camelCase`.
- Packages: short, lowercase names (`handler`, `serverStub`). Prefer putting new entrypoints under `cmd/<name>/main.go` to avoid multiple `main()` collisions in the root package.
- Python demos: follow basic PEP 8 and keep scripts dependency-light.

## Testing Guidelines

- Go tests should use the standard `testing` package and live next to code as `*_test.go`.
- Run tests with `go test ./...`. If the root package fails to build due to multiple `main()` files, move programs into `cmd/` first.

## Commit & Pull Request Guidelines

- Git history is not established yet; use clear, imperative messages. Suggested pattern: `feat: ...`, `fix: ...`, `chore: ...`.
- PRs should include: a short description, how to run/verify (exact command), and any port/protocol changes (e.g., `:1234`, `:8082`).

## Security & Configuration Notes

- Servers currently bind to local ports without auth; do not expose them publicly without adding access controls and timeouts.
