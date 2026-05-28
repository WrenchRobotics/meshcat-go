package meshcat

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/WrenchRobotics/meshcat-go/geometry"
	zmqserver "github.com/WrenchRobotics/meshcat-go/servers/zmq"
	"github.com/ugorji/go/codec"
	"github.com/zeromq/goczmq"
)

// ViewerWindow manages a Meshcat viewer window and its connection.
// It abstracts away the details of starting the server, managing connections,
// and opening the viewer in a browser.
type ViewerWindow struct {
	app       *MeshcatWebServerApplication
	zmqURL    string
	webURL    string
	client    MeshcatClient
	running   bool
	mutex     sync.Mutex
	zmqSocket *goczmq.Sock
}

const defaultWebsocketConnectTimeout = 10 * time.Second

// NewViewerWindow creates and starts a new local Meshcat viewer window using
// an available HTTP port on the local machine.
func NewViewerWindow() (*ViewerWindow, error) {
	app := NewMeshcatWebServerApplication()
	return newViewerWindowFromApp(app), nil
}

// NewViewerWindowWithPort creates and starts a new local Meshcat viewer window
// using the provided web server port.
//
// This value is forwarded to MeshcatWebServerApplication and then into the
// ZeroMQWebsocketBridge so that the viewer URL and HTTP listener port stay in
// sync.
func NewViewerWindowWithPort(webPort int) (*ViewerWindow, error) {
	app := NewMeshcatWebServerApplicationWithPort(webPort)
	return newViewerWindowFromApp(app), nil
}

func newViewerWindowFromApp(app MeshcatWebServerApplication) *ViewerWindow {
	// Start the server
	(&app).Start()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	zmqURL := app.Bridge.ZMQUrl
	webURL := app.Bridge.WebUrl

	// Create a ZMQ client connected to the bridge
	client := &ZMQClient{
		zmqURL: zmqURL,
		bridge: app.Bridge,
	}

	window := &ViewerWindow{
		app:     &app,
		zmqURL:  zmqURL,
		webURL:  webURL,
		client:  client,
		running: true,
	}

	log.Printf("ViewerWindow started: %s", webURL)

	return window
}

// NewViewerWindowRemote connects to a remote Meshcat viewer at the given URL.
func NewViewerWindowRemote(webURL string, zmqURL string) (*ViewerWindow, error) {
	// Create a ZMQ client connected to the remote bridge
	client := &ZMQClient{
		zmqURL: zmqURL,
		bridge: nil,
	}

	window := &ViewerWindow{
		app:     nil,
		zmqURL:  zmqURL,
		webURL:  webURL,
		client:  client,
		running: true,
	}

	log.Printf("ViewerWindow connected to remote: %s", webURL)

	return window, nil
}

// URL returns the web URL of the Meshcat viewer.
func (w *ViewerWindow) URL() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.webURL
}

// ZMQURL returns the ZMQ URL for communication with the server.
func (w *ViewerWindow) ZMQURL() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.zmqURL
}

// Open opens the Meshcat viewer in the default system browser.
func (w *ViewerWindow) Open() error {
	w.mutex.Lock()
	url := w.webURL
	app := w.app
	w.mutex.Unlock()

	if err := openInBrowser(url); err != nil {
		return err
	}

	if app != nil && app.Bridge != nil {
		if err := app.Bridge.WaitForWebsocketConnection(defaultWebsocketConnectTimeout); err != nil {
			return err
		}
	}

	return nil
}

// Client returns the underlying MeshcatClient.
func (w *ViewerWindow) Client() MeshcatClient {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.client
}

// Visualizer returns a new Visualizer connected to this window.
func (w *ViewerWindow) Visualizer() *Visualizer {
	return NewVisualizer(w.client)
}

// Stop stops the Meshcat server.
func (w *ViewerWindow) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.app != nil {
		if err := w.app.Stop(); err != nil {
			return err
		}
		w.running = false
	}

	return nil
}

// Close stops the server and cleans up resources.
func (w *ViewerWindow) Close() error {
	return w.Stop()
}

// ZMQClient implements MeshcatClient using ZeroMQ.
type ZMQClient struct {
	zmqURL string
	bridge *zmqserver.ZeroMQWebsocketBridge
	mutex  sync.Mutex
}

type lowerableSceneObject interface {
	Lower() map[string]any
}

// SetObject sends a set_object command to the Meshcat server.
func (zc *ZMQClient) SetObject(path []string, obj interface{}) error {
	normalizedObj := normalizeSetObjectPayload(obj)
	cmd, err := zc.makeSetObjectCommand(path, normalizedObj)
	if err != nil {
		return fmt.Errorf("failed to build set_object command: %w", err)
	}
	if zc.bridge != nil {
		// Direct bridge access for local servers
		frames := [][]byte{[]byte("set_object"), []byte(formatPath(path)), cmd}
		zc.bridge.HandleLocalFrameSlice(frames)
		return nil
	}

	// Remote connection via ZMQ
	return zc.sendZMQCommand("set_object", path, cmd)
}

func normalizeSetObjectPayload(obj interface{}) interface{} {
	if loweredObj, ok := obj.(lowerableSceneObject); ok {
		return loweredObj.Lower()
	}

	if geom, ok := obj.(geometry.Geometry); ok {
		// Match meshcat-python semantics: plain geometry is wrapped in a Mesh
		// with a default MeshPhongMaterial.
		return geometry.NewMesh(geom, geometry.NewMeshPhongMaterial()).Lower()
	}

	return obj
}

// SetTransform sends a set_transform command to the Meshcat server.
func (zc *ZMQClient) SetTransform(path []string, transform [4][4]float64) error {
	cmd, err := zc.makeSetTransformCommand(path, transform)
	if err != nil {
		return fmt.Errorf("failed to build set_transform command: %w", err)
	}
	if zc.bridge != nil {
		// Direct bridge access for local servers
		frames := [][]byte{[]byte("set_transform"), []byte(formatPath(path)), cmd}
		zc.bridge.HandleLocalFrameSlice(frames)
		return nil
	}

	// Remote connection via ZMQ
	return zc.sendZMQCommand("set_transform", path, cmd)
}

// SetProperty sends a set_property command to the Meshcat server.
func (zc *ZMQClient) SetProperty(path []string, property string, value interface{}) error {
	cmd, err := zc.makeSetPropertyCommand(path, property, value)
	if err != nil {
		return fmt.Errorf("failed to build set_property command: %w", err)
	}
	if zc.bridge != nil {
		frames := [][]byte{[]byte("set_property"), []byte(formatPath(path)), cmd}
		zc.bridge.HandleLocalFrameSlice(frames)
		return nil
	}

	return zc.sendZMQCommand("set_property", path, cmd)
}

// Delete sends a delete command to the Meshcat server.
func (zc *ZMQClient) Delete(path []string) error {
	cmd, err := zc.makeDeleteCommand(path)
	if err != nil {
		return fmt.Errorf("failed to build delete command: %w", err)
	}
	if zc.bridge != nil {
		// Direct bridge access for local servers
		frames := [][]byte{[]byte("delete"), []byte(formatPath(path)), cmd}
		zc.bridge.HandleLocalFrameSlice(frames)
		return nil
	}

	// Remote connection via ZMQ
	return zc.sendZMQCommand("delete", path, cmd)
}

// Open opens the viewer in a browser.
func (zc *ZMQClient) Open() error {
	// This is handled by ViewerWindow.Open()
	return nil
}

// Helper methods

func (zc *ZMQClient) makeSetObjectCommand(path []string, obj interface{}) ([]byte, error) {
	cmd := map[string]interface{}{
		"type":   "set_object",
		"path":   formatPath(path),
		"object": obj,
	}
	return zc.msgpackEncode(cmd)
}

func (zc *ZMQClient) makeSetTransformCommand(path []string, transform [4][4]float64) ([]byte, error) {
	// Convert transform matrix to flat array for msgpack
	matrix := make([]float64, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			matrix[i*4+j] = transform[i][j]
		}
	}

	cmd := map[string]interface{}{
		"type":   "set_transform",
		"path":   formatPath(path),
		"matrix": matrix,
	}
	return zc.msgpackEncode(cmd)
}

func (zc *ZMQClient) makeSetPropertyCommand(path []string, property string, value interface{}) ([]byte, error) {
	cmd := map[string]interface{}{
		"type":     "set_property",
		"path":     formatPath(path),
		"property": property,
		"value":    value,
	}
	return zc.msgpackEncode(cmd)
}

func (zc *ZMQClient) makeDeleteCommand(path []string) ([]byte, error) {
	cmd := map[string]interface{}{
		"type": "delete",
		"path": formatPath(path),
	}
	return zc.msgpackEncode(cmd)
}

func (zc *ZMQClient) sendZMQCommand(cmd string, path []string, payload []byte) error {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	req := goczmq.NewSock(goczmq.Req)
	if req == nil {
		return fmt.Errorf("failed to create ZMQ REQ socket")
	}
	defer req.Destroy()

	if err := req.Connect(zc.zmqURL); err != nil {
		return fmt.Errorf("failed to connect to ZMQ: %v", err)
	}

	pathStr := formatPath(path)

	// Send command, path, and data as three frames
	if err := req.SendFrame([]byte(cmd), goczmq.FlagMore); err != nil {
		return err
	}
	if err := req.SendFrame([]byte(pathStr), goczmq.FlagMore); err != nil {
		return err
	}

	if err := req.SendFrame(payload, goczmq.FlagNone); err != nil {
		return err
	}

	// Wait for reply
	reply, _, err := req.RecvFrame()
	if err != nil {
		return err
	}

	replyText := strings.TrimSpace(string(reply))
	if replyText != "ok" {
		if replyText == "" {
			replyText = "empty reply"
		}
		return fmt.Errorf("zmq command %q failed: %s", cmd, replyText)
	}

	return nil
}

func (zc *ZMQClient) msgpackEncode(v interface{}) ([]byte, error) {
	var out []byte
	h := new(codec.MsgpackHandle)
	enc := codec.NewEncoderBytes(&out, h)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to msgpack encode: %w", err)
	}
	return out, nil
}

func formatPath(path []string) string {
	if len(path) == 0 {
		return "/"
	}

	return "/" + strings.Join(path, "/")
}

// openInBrowser opens the given URL in the system default browser.
func openInBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
