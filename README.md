# meshcat-go
A library defined to wrap around the MeshCat 3D Viewer in Golang.

## Installation

This project uses `github.com/zeromq/goczmq` for the ZeroMQ bridge.
On macOS, install the native libraries first:

```bash
brew install libsodium zeromq czmq
```

To add this library to your own Go module:

```bash
go get github.com/WrenchRobotics/meshcat-go@latest
```

If you cloned this repository for local development, fetch Go dependencies:

```bash
go mod download
```

Notes:

- These native libraries are required for the ZeroMQ bridge and examples (including `examples/get_scene`).
- If you only consume packages that do not touch the ZMQ bridge, you may not need them at runtime, but building/testing bridge-related code will require them.

## MeshCat Viewer Submodule

This repository uses a git submodule to pin the upstream JavaScript viewer:

- Repository: `https://github.com/meshcat-dev/meshcat`
- Commit: `65781fcb064db536b99a66fe9fcf5bf0b6d1f790`
- Path: `third_party/meshcat-js`

When cloning this repository, initialize submodules:

```bash
git submodule update --init --recursive
```

The Go websocket app serves the pinned viewer files from:

- `http://<host>:<port>/viewer/`

The viewer's JavaScript client connects to the websocket endpoint at `ws://<host>:<port>/ws`.