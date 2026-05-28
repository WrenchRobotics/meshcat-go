# zmqserver1 example

This example demonstrates direct usage of the Meshcat ZeroMQ server protocol from Go.

It starts the Meshcat web server + ZMQ bridge, then uses a ZMQ REQ client to issue commands and read replies.

## Features shown

- `url`: ask the bridge for the viewer URL.
- `wait` (optional): block until a websocket client connects.
- `set_object`: add a box mesh at a scene path.
- `set_transform`: move the object with a transform matrix.
- `set_property`: modify object appearance (`opacity` in this example).
- `get_scene`: request a self-contained HTML snapshot of the current scene.
- `delete`: remove the object path from the scene tree.

## Prerequisites

Install native dependencies required by `goczmq`:

```bash
brew install libsodium zeromq czmq
```

Install Go dependencies from repository root:

```bash
go mod download
```

## Run

From repository root:

```bash
go run ./examples/zmqserver1
```

### Useful flags

Do not auto-open the browser:

```bash
go run ./examples/zmqserver1 -open-viewer=false
```

Also issue the `wait` command (requires browser websocket connection):

```bash
go run ./examples/zmqserver1 -wait-for-websocket
```

## Expected output

You should see logs similar to:

```text
Bridge ZMQ URL: tcp://127.0.0.1:6000
Viewer URL: http://127.0.0.1:7xxx/static/
url command reply: http://127.0.0.1:7xxx/static/
get_scene wrote HTML to: /var/folders/.../meshcat-zmqserver1-XXXXXX.html
ZMQ feature demo complete
```

## Notes

- ZMQ REQ/REP requires strict send-then-receive ordering for each command.
- Meshcat commands (`set_object`, `set_transform`, etc.) are sent as 3 frames: `command`, `path`, `msgpack_payload`.
- `get_scene` returns HTML bytes directly as the reply payload.