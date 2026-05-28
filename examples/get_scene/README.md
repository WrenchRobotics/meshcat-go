# get_scene example

This example demonstrates the ZeroMQ `get_scene` request/response flow end-to-end.

It does four things:

1. Starts a `ZeroMQWebsocketBridge`.
2. Pre-populates the bridge scene tree with a simple box object and transform.
3. Connects as a ZeroMQ REQ client and sends the `get_scene` command.
4. Receives the HTML scene payload, writes it to a temporary `.html` file, and opens it in your default browser.

## Why this example exists

Use this example to verify that:

- bridge startup works,
- scene tree data is serialized correctly,
- `get_scene` returns valid HTML,
- and the generated page renders in a browser.

## Prerequisites

Install native dependencies used by CZMQ/ZeroMQ:

```bash
brew install libsodium zeromq czmq
```

From the repository root, ensure Go dependencies are available:

```bash
go mod download
```

## Run

From the repository root:

```bash
go run ./examples/get_scene
```

Or from this folder:

```bash
go run main.go
```

## Expected output

You should see logs similar to:

```text
Bridge ZMQ URL: tcp://127.0.0.1:6000
Received HTML reply (N bytes)
Scene HTML written to: /var/folders/.../meshcat-scene-XXXXXX.html
```

Then your default browser should open a local file URL (`file://...`) containing the generated MeshCat scene.

## Notes

- The HTML file is created in your system temp directory.
- If automatic browser launch fails, copy the printed path and open it manually.
- The example is intentionally minimal and is useful as a smoke test for scene export.
