# meshcat-go
A library defined to wrap around the MeshCat 3D Viewer in Golang.

## Installation

### Prerequisites

This project uses `github.com/zeromq/goczmq` for the ZeroMQ bridge.
On macOS, install the native libraries first:

```bash
brew install libsodium zeromq czmq
```

### Installation (Standard)

To add this library to your own Go module:

```bash
go get github.com/WrenchRobotics/meshcat-go@latest
```

### Local Installation (Advanced, For Developers)

If you cloned this repository for local development, fetch Go dependencies:

```bash
go mod download
```

Notes:

- These native libraries are required for the ZeroMQ bridge and examples (including `examples/get_scene`).
- If you only consume packages that do not touch the ZMQ bridge, you may not need them at runtime, but building/testing bridge-related code will require them.

## MeshCat Viewer Assets

The built MeshCat viewer files are vendored directly in this repository at:

- `viewer_assets/dist`

These files are embedded into the Go binary, so standard source checkouts and CI jobs do not need git submodules to serve the viewer.

### Updating Vendored Viewer Assets (Maintainers)

When you need to refresh to a newer upstream MeshCat viewer build, replace the files in `viewer_assets/dist` with a new `dist` build output.
The required files are:

- `index.html`
- `main.min.js`
- `main.min.js.THIRD_PARTY_LICENSES.json`

The Go websocket app serves the vendored viewer files from:

- `http://<host>:<port>/static/`

The viewer's JavaScript client connects to the websocket endpoint at `ws://<host>:<port>/ws`.