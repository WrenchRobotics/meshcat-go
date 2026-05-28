// zmqserver1 demonstrates direct use of the Meshcat ZeroMQ server protocol
// with a REQ client.
//
// Run from repository root:
//
//	go run ./examples/zmqserver1
//
// Optional flags:
//
//	go run ./examples/zmqserver1 -open-viewer=false
//	go run ./examples/zmqserver1 -wait-for-websocket
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/WrenchRobotics/meshcat-go/geometry"
	"github.com/WrenchRobotics/meshcat-go/meshcat"
	"github.com/ugorji/go/codec"
	"github.com/zeromq/goczmq"
)

const requestTimeout = 5 * time.Second

// msgpackEncode serializes a Go value into msgpack bytes, which is the wire
// format expected by Meshcat commands sent over ZMQ.
func msgpackEncode(v any) ([]byte, error) {
	var out []byte
	h := new(codec.MsgpackHandle)
	enc := codec.NewEncoderBytes(&out, h)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return out, nil
}

// buildSetObjectCommandPayload constructs a set_object payload containing a
// simple colored box mesh.
func buildSetObjectCommandPayload(path string) ([]byte, error) {
	mat := geometry.NewMeshLambertMaterial()
	mat.Color = 0x1f7a8c
	mesh := geometry.NewMesh(geometry.NewBox([3]float64{0.4, 0.4, 0.4}), mat)

	cmd := map[string]any{
		"type":   "set_object",
		"path":   path,
		"object": mesh.Lower(),
	}

	return msgpackEncode(cmd)
}

// buildSetTransformCommandPayload constructs a set_transform payload that
// translates the object in the scene.
func buildSetTransformCommandPayload(path string, x, y, z float64) ([]byte, error) {
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

// buildSetPropertyCommandPayload constructs a set_property payload used here to
// set object opacity.
func buildSetPropertyCommandPayload(path string, property string, value any) ([]byte, error) {
	cmd := map[string]any{
		"type":     "set_property",
		"path":     path,
		"property": property,
		"value":    value,
	}

	return msgpackEncode(cmd)
}

// buildDeleteCommandPayload constructs a delete payload for a scene path.
func buildDeleteCommandPayload(path string) ([]byte, error) {
	cmd := map[string]any{
		"type": "delete",
		"path": path,
	}

	return msgpackEncode(cmd)
}

// sendSingleFrameCommand sends a one-frame command (for example: url, wait,
// get_scene) and returns the first reply frame.
func sendSingleFrameCommand(req *goczmq.Sock, command string) ([]byte, error) {
	if err := req.SendFrame([]byte(command), goczmq.FlagNone); err != nil {
		return nil, fmt.Errorf("send %q failed: %w", command, err)
	}

	reply, err := req.RecvMessage()
	if err != nil {
		return nil, fmt.Errorf("receive reply for %q failed: %w", command, err)
	}
	if len(reply) == 0 {
		return nil, fmt.Errorf("empty reply for %q", command)
	}

	return reply[0], nil
}

// sendMeshcatCommand sends a three-frame Meshcat command and validates that
// the server acknowledged the request with "ok".
func sendMeshcatCommand(req *goczmq.Sock, command, path string, payload []byte) error {
	if err := req.SendFrame([]byte(command), goczmq.FlagMore); err != nil {
		return fmt.Errorf("send command frame failed: %w", err)
	}
	if err := req.SendFrame([]byte(path), goczmq.FlagMore); err != nil {
		return fmt.Errorf("send path frame failed: %w", err)
	}
	if err := req.SendFrame(payload, goczmq.FlagNone); err != nil {
		return fmt.Errorf("send payload frame failed: %w", err)
	}

	reply, err := req.RecvMessage()
	if err != nil {
		return fmt.Errorf("receive reply for %q failed: %w", command, err)
	}
	if len(reply) == 0 {
		return errors.New("empty reply from server")
	}
	if string(reply[0]) != "ok" {
		return fmt.Errorf("unexpected reply for %q: %q", command, string(reply[0]))
	}

	return nil
}

// writeSceneHTML writes get_scene output to a temporary HTML file and returns
// the file path.
func writeSceneHTML(html []byte) (string, error) {
	f, err := os.CreateTemp("", "meshcat-zmqserver1-*.html")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(html); err != nil {
		return "", err
	}

	return f.Name(), nil
}

// openInBrowser best-effort opens a URL in the system default browser.
func openInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %q", runtime.GOOS)
	}

	return cmd.Start()
}

func main() {
	openViewer := flag.Bool("open-viewer", true, "open Meshcat viewer URL in a browser")
	waitForWebsocket := flag.Bool("wait-for-websocket", false, "send the wait command (requires a websocket client)")
	flag.Parse()

	// Start an application that runs both the HTTP viewer and ZMQ bridge.
	app := meshcat.NewMeshcatWebServerApplication()
	app.Start()
	defer func() {
		if err := app.Stop(); err != nil {
			log.Printf("app shutdown error: %v", err)
		}
	}()

	// Allow listeners to begin accepting connections before creating the REQ
	// socket.
	time.Sleep(100 * time.Millisecond)

	log.Printf("Bridge ZMQ URL: %s", app.Bridge.ZMQUrl)
	log.Printf("Viewer URL: %s", app.Bridge.WebUrl)

	if *openViewer {
		if err := openInBrowser(app.Bridge.WebUrl); err != nil {
			log.Printf("could not open browser automatically: %v", err)
		}
	}

	req := goczmq.NewSock(goczmq.Req)
	if req == nil {
		log.Fatal("failed to create ZMQ REQ socket")
	}
	defer req.Destroy()

	if err := req.Connect(app.Bridge.ZMQUrl); err != nil {
		log.Fatalf("failed to connect REQ socket: %v", err)
	}
	req.SetRcvtimeo(int(requestTimeout.Milliseconds()))

	urlReply, err := sendSingleFrameCommand(req, "url")
	if err != nil {
		log.Fatalf("url command failed: %v", err)
	}
	log.Printf("url command reply: %s", string(urlReply))

	if *waitForWebsocket {
		waitReply, err := sendSingleFrameCommand(req, "wait")
		if err != nil {
			log.Fatalf("wait command failed: %v", err)
		}
		log.Printf("wait command reply: %s", string(waitReply))
	}

	objectPath := "/meshcat/box1"

	setObjectPayload, err := buildSetObjectCommandPayload(objectPath)
	if err != nil {
		log.Fatalf("failed to build set_object payload: %v", err)
	}
	if err := sendMeshcatCommand(req, "set_object", objectPath, setObjectPayload); err != nil {
		log.Fatalf("set_object failed: %v", err)
	}

	setTransformPayload, err := buildSetTransformCommandPayload(objectPath, 0.0, 0.0, 0.25)
	if err != nil {
		log.Fatalf("failed to build set_transform payload: %v", err)
	}
	if err := sendMeshcatCommand(req, "set_transform", objectPath, setTransformPayload); err != nil {
		log.Fatalf("set_transform failed: %v", err)
	}

	setPropertyPayload, err := buildSetPropertyCommandPayload(objectPath, "opacity", 0.8)
	if err != nil {
		log.Fatalf("failed to build set_property payload: %v", err)
	}
	if err := sendMeshcatCommand(req, "set_property", objectPath, setPropertyPayload); err != nil {
		log.Fatalf("set_property failed: %v", err)
	}

	getSceneReply, err := sendSingleFrameCommand(req, "get_scene")
	if err != nil {
		log.Fatalf("get_scene failed: %v", err)
	}

	htmlPath, err := writeSceneHTML(getSceneReply)
	if err != nil {
		log.Fatalf("failed to write get_scene HTML: %v", err)
	}
	log.Printf("get_scene wrote HTML to: %s", htmlPath)

	deletePayload, err := buildDeleteCommandPayload(objectPath)
	if err != nil {
		log.Fatalf("failed to build delete payload: %v", err)
	}
	if err := sendMeshcatCommand(req, "delete", objectPath, deletePayload); err != nil {
		log.Fatalf("delete failed: %v", err)
	}

	log.Println("ZMQ feature demo complete")
	log.Println("Tip: open the generated HTML file to inspect the serialized scene snapshot")
}
