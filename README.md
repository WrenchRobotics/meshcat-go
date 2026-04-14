# meshcat-go
A library defined to wrap around the MeshCat 3D Viewer in Golang.

## Installation

```bash
brew install libsodium zeromq czmq
```

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