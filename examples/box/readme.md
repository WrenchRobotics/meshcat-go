# Meshcat Go Example: Box

This example demonstrates how to use the MeshcatVisualizer and ViewerWindow interfaces in Go to create and manipulate 3D objects in a Meshcat viewer, similar to the Python `meshcat` library's `box.py` example.

## What it does
- Creates and starts a local Meshcat server (ViewerWindow)
- Connects to it using the Go MeshcatVisualizer
- Opens the Meshcat viewer and waits for the browser websocket to connect
- Adds a box at the root of the scene
- Applies a transform to the box
- Adds a second, smaller box at `/second` and applies a different transform

## How to run

1. Make sure you have all dependencies installed:
   ```sh
   go mod tidy
   ```
2. Run the example:
   ```sh
   go run ./examples/box
   ```
3. The Meshcat viewer should open automatically in your browser. If not, open the printed URL manually.

## How it works
- The example creates a `ViewerWindow`, which handles starting the Meshcat server and managing connections.
- You call `window.Visualizer()` to get a `Visualizer` that lets you manipulate the scene.
- The `Visualizer` API mirrors the Python meshcat library, allowing you to chain `.At("path")` calls and use `.SetObject`, `.SetTransform`, and `.Delete` to manipulate objects.

## Example code
```go
window, err := meshcat.NewViewerWindow()
defer window.Close()

v := window.Visualizer()
window.Open()
v.SetObject(geometry.NewBox([3]float64{1, 1, 1}))
v.SetTransform(transform)
```

## Notes
- The ViewerWindow automatically manages the Meshcat server lifecycle.
- For remote servers, use `meshcat.NewViewerWindowRemote(webURL, zmqURL)` instead.
- For more advanced usage, see the Python meshcat documentation and adapt the API to Go idioms.
