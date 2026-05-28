// get_scene is a manual end-to-end example that:
//
//  1. Starts a ZeroMQWebsocketBridge with a pre-populated scene tree.
//  2. Connects to it as a ZMQ REQ client and sends a "get_scene" command.
//  3. Writes the HTML reply to a temporary file.
//  4. Opens the file in the default browser so you can verify the viewer loads.
//
// Run with:
//
//	go run ./examples/get_scene
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	zmqserver "github.com/WrenchRobotics/meshcat-go/servers/zmq"
	"github.com/ugorji/go/codec"
	"github.com/zeromq/goczmq"
)

func msgpackEncode(v any) ([]byte, error) {
	var out []byte
	h := new(codec.MsgpackHandle)
	enc := codec.NewEncoderBytes(&out, h)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to msgpack encode command: %w", err)
	}
	return out, nil
}

func makeSetObjectBoxCmd(path string) ([]byte, error) {
	geometryUUID := "box-geometry"
	materialUUID := "box-material"
	objectUUID := "box-object"

	cmd := map[string]any{
		"type": "set_object",
		"path": path,
		"object": map[string]any{
			"metadata": map[string]any{
				"version": 4.5,
				"type":    "Object",
			},
			"geometries": []any{
				map[string]any{
					"uuid":   geometryUUID,
					"type":   "BoxGeometry",
					"width":  0.4,
					"height": 0.4,
					"depth":  0.4,
				},
			},
			"materials": []any{
				map[string]any{
					"uuid":         materialUUID,
					"type":         "MeshLambertMaterial",
					"color":        0x33aaee,
					"reflectivity": 0.5,
					"wireframe":    false,
				},
			},
			"object": map[string]any{
				"uuid":     objectUUID,
				"type":     "Mesh",
				"geometry": geometryUUID,
				"material": materialUUID,
				"matrix": []float64{
					1, 0, 0, 0,
					0, 1, 0, 0,
					0, 0, 1, 0,
					0, 0, 0, 1,
				},
			},
		},
	}

	return msgpackEncode(cmd)
}

func makeSetTransformCmd(path string, x, y, z float64) ([]byte, error) {
	cmd := map[string]any{
		"type": "set_transform",
		"path": path,
		"matrix": []float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			x, y, z, 1,
		},
	}
	return msgpackEncode(cmd)
}

func main() {
	// ── 1. Start the bridge ───────────────────────────────────────────────────
	bridge := zmqserver.NewZeroMQWebsocketBridge()

	// Populate the scene tree with valid MeshCat msgpack commands so the
	// generated HTML contains visible geometry.
	box := bridge.SceneTree.GetPath([]string{"meshcat", "box"})
	boxPath := "/meshcat/box"
	boxObjectCmd, err := makeSetObjectBoxCmd(boxPath)
	if err != nil {
		log.Fatalf("failed to create set_object command: %v", err)
	}
	boxTransformCmd, err := makeSetTransformCmd(boxPath, 0.0, 0.0, 0.2)
	if err != nil {
		log.Fatalf("failed to create set_transform command: %v", err)
	}
	box.Object = boxObjectCmd
	box.Transform = boxTransformCmd

	go bridge.Run()
	defer bridge.Stop()

	// Give the run loop a moment to start polling.
	time.Sleep(50 * time.Millisecond)

	log.Printf("Bridge ZMQ URL: %s", bridge.ZMQUrl)

	// ── 2. Connect as a REQ client and send "get_scene" ───────────────────────
	req := goczmq.NewSock(goczmq.Req)
	if req == nil {
		log.Fatal("failed to create REQ socket")
	}
	defer req.Destroy()

	if err := req.Connect(bridge.ZMQUrl); err != nil {
		log.Fatalf("failed to connect to bridge: %v", err)
	}

	if err := req.SendFrame([]byte("get_scene"), goczmq.FlagNone); err != nil {
		log.Fatalf("failed to send get_scene: %v", err)
	}

	// ── 3. Receive the HTML reply ─────────────────────────────────────────────
	req.SetRcvtimeo(5000)
	reply, err := req.RecvMessage()
	if err != nil {
		log.Fatalf("failed to receive get_scene reply: %v", err)
	}
	if len(reply) == 0 {
		log.Fatal("received empty reply")
	}
	html := reply[0]
	log.Printf("Received HTML reply (%d bytes)", len(html))

	// ── 4. Write to a temp file and open in the browser ───────────────────────
	f, err := os.CreateTemp("", "meshcat-scene-*.html")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.Write(html); err != nil {
		f.Close()
		log.Fatalf("failed to write HTML: %v", err)
	}
	f.Close()

	path := f.Name()
	fmt.Printf("Scene HTML written to: %s\n", path)
	openInBrowser("file://" + path)
}

// openInBrowser opens the given URL in the system default browser.
func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		fmt.Printf("Open this file in your browser: %s\n", url)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Could not open browser automatically. Open this file manually: %s\n", url)
	}
}
