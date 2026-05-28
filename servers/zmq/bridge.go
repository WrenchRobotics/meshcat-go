package zmq

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	meshcatgo "github.com/WrenchRobotics/meshcat-go"
	"github.com/WrenchRobotics/meshcat-go/commands"
	"github.com/WrenchRobotics/meshcat-go/scene"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zeromq/goczmq"
)

// ZeroMQWebsocketBridge forwards Meshcat commands between a ZMQ REP socket and
// browser websocket clients, while maintaining an in-memory scene cache.
type ZeroMQWebsocketBridge struct {
	GorillaUpgrader *websocket.Upgrader
	ZMQUrl          string
	WebUrl          string // The URL where the websocket can be accessed by clients (e.g. the Meshcat viewer). This is typically different from ZMQUrl since the websocket is served via the Gin router on a different port.
	Host            string
	Port            int
	CertificateFile string
	KeyFile         string
	ZMQStream       *goczmq.Sock
	SceneTree       *scene.TreeNode

	// Hidden fields for internal use

	// wsMu is used to synchronize access to the wsPool, which can be accessed concurrently by multiple goroutines (e.g. the main Run loop and the WebsocketHandler).
	wsMu sync.RWMutex

	// Websocket connection pool to keep track of all active websocket connections to the client.
	wsPool map[*websocket.Conn]struct{}

	// stopCh is used to signal the Run loop to stop.
	// It should be closed when we want to stop the bridge.
	stopCh chan struct{}

	// stopOnce is used to ensure that the stopCh is only closed once,
	// even if Stop is called multiple times.
	stopOnce sync.Once

	// captureMu guards captureRespCh lifecycle.
	captureMu sync.Mutex

	// captureRespCh is non-nil while waiting for a capture_image response
	// from a websocket client.
	captureRespCh chan []byte
}

func NewZeroMQServer(
	url string,
	host string,
	port int,
	certificateFile string,
	keyFile string,
	ngrokHttpTunnel bool,
) *ZeroMQWebsocketBridge {

	return &ZeroMQWebsocketBridge{
		ZMQUrl:          url,
		Host:            host,
		Port:            port,
		CertificateFile: certificateFile,
		KeyFile:         keyFile,
		wsPool:          make(map[*websocket.Conn]struct{}),
		stopCh:          make(chan struct{}),
	}
}

func (bridge *ZeroMQWebsocketBridge) Destroy() {
	if bridge.ZMQStream != nil {
		bridge.ZMQStream.Destroy()
	}
}

func (bridge *ZeroMQWebsocketBridge) addWebsocketConn(conn *websocket.Conn) {
	bridge.wsMu.Lock()
	defer bridge.wsMu.Unlock()

	if bridge.wsPool == nil {
		bridge.wsPool = make(map[*websocket.Conn]struct{})
	}

	bridge.wsPool[conn] = struct{}{}
	if conn != nil {
		log.Printf("WebSocket connection added: %v. Total connections: %d", conn.RemoteAddr(), len(bridge.wsPool))
	} else {
		log.Printf("WebSocket connection added: <nil>. Total connections: %d", len(bridge.wsPool))
	}
}

func (bridge *ZeroMQWebsocketBridge) removeWebsocketConn(conn *websocket.Conn) {
	bridge.wsMu.Lock()
	defer bridge.wsMu.Unlock()

	if bridge.wsPool != nil {
		delete(bridge.wsPool, conn)
		if conn != nil {
			log.Printf("WebSocket connection removed: %v. Total connections: %d", conn.RemoteAddr(), len(bridge.wsPool))
		} else {
			log.Printf("WebSocket connection removed: <nil>. Total connections: %d", len(bridge.wsPool))
		}
	}
}

func (bridge *ZeroMQWebsocketBridge) hasWebsocketConn() bool {
	bridge.wsMu.RLock()
	defer bridge.wsMu.RUnlock()

	return len(bridge.wsPool) > 0
}

func (bridge *ZeroMQWebsocketBridge) forwardToWebsockets(data []byte) {
	bridge.wsMu.RLock()
	if len(bridge.wsPool) == 0 {
		bridge.wsMu.RUnlock()
		log.Println("No active WebSocket connections. Cannot forward data.")
		return
	}
	conns := make([]*websocket.Conn, 0, len(bridge.wsPool))
	for conn := range bridge.wsPool {
		if conn != nil {
			conns = append(conns, conn)
		}
	}
	bridge.wsMu.RUnlock()

	log.Printf("Forwarding data to WebSocket connections: %d bytes", len(data))
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			log.Printf("Failed to forward data to WebSocket: %v", err)
		}
	}
}

func (bridge *ZeroMQWebsocketBridge) sendScene(conn *websocket.Conn) {
	if conn == nil || bridge.SceneTree == nil {
		return
	}

	bridge.SceneTree.Walk(func(node *scene.TreeNode) {
		if node.Object != nil {
			if b, ok := node.Object.([]byte); ok {
				if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					log.Printf("failed to send object to websocket: %v", err)
				}
			}
		}
		for _, p := range node.Properties {
			if b, ok := p.([]byte); ok {
				if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					log.Printf("failed to send property to websocket: %v", err)
				}
			}
		}
		if node.Transform != nil {
			if b, ok := node.Transform.([]byte); ok {
				if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					log.Printf("failed to send transform to websocket: %v", err)
				}
			}
		}
		if node.Animation != nil {
			if b, ok := node.Animation.([]byte); ok {
				if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					log.Printf("failed to send animation to websocket: %v", err)
				}
			}
		}
	})
}

func (bridge *ZeroMQWebsocketBridge) waitForWebsockets() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if bridge.hasWebsocketConn() {
			if err := bridge.ZMQStream.SendFrame([]byte("ok"), goczmq.FlagNone); err != nil {
				log.Printf("failed to reply to wait command: %v", err)
			}
			return
		}

		select {
		case <-bridge.stopCh:
			return
		case <-ticker.C:
		}
	}
}

func (bridge *ZeroMQWebsocketBridge) WaitForWebsocketConnection(timeout time.Duration) error {
	if bridge == nil {
		return fmt.Errorf("bridge is nil")
	}

	if bridge.hasWebsocketConn() {
		return nil
	}

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		if bridge.hasWebsocketConn() {
			return nil
		}

		select {
		case <-bridge.stopCh:
			return fmt.Errorf("bridge stopped before websocket connected")
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for websocket connection")
		case <-ticker.C:
		}
	}
}

func decodeCaptureImagePayload(msg []byte) ([]byte, error) {
	var payload struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(msg, &payload); err != nil {
		return nil, err
	}

	parts := strings.SplitN(payload.Data, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("capture payload missing data url prefix")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func (bridge *ZeroMQWebsocketBridge) setCaptureResponseChannel(ch chan []byte) {
	bridge.captureMu.Lock()
	defer bridge.captureMu.Unlock()
	bridge.captureRespCh = ch
}

func (bridge *ZeroMQWebsocketBridge) getCaptureResponseChannel() chan []byte {
	bridge.captureMu.Lock()
	defer bridge.captureMu.Unlock()
	return bridge.captureRespCh
}

func (bridge *ZeroMQWebsocketBridge) sendExpected3FramesError(cmd string) {
	if err := bridge.ZMQStream.SendFrame([]byte("error: expected 3 frames"), goczmq.FlagNone); err != nil {
		log.Printf("failed to send frame-count error for command %s: %v", cmd, err)
	}
}

// HandleZMQFrameSlice is a helper function that takes in a slice of ZMQ frames and
// handles the input frames according to Meshcat's ZMQ protocol.
// It returns a string response that can be sent back to the client.
func (bridge *ZeroMQWebsocketBridge) HandleZMQFrameSlice(frames [][]byte) {
	bridge.handleFrameSlice(frames, true)
}

func (bridge *ZeroMQWebsocketBridge) HandleLocalFrameSlice(frames [][]byte) {
	bridge.handleFrameSlice(frames, false)
}

func (bridge *ZeroMQWebsocketBridge) handleFrameSlice(frames [][]byte, sendReplies bool) {
	// Input Checking
	// Input Checking
	if len(frames) == 0 {
		log.Println("Received empty frame slice")
		if sendReplies {
			_ = bridge.ZMQStream.SendFrame([]byte("error: empty request"), goczmq.FlagNone)
		}
		return
	}

	// Attempt to handle the frames according to:
	// - Standard ZMQ commands (e.g. "url", "wait", etc.)
	cmd := string(frames[0])
	log.Printf("Handling ZMQ command: %s", cmd)

	if sendReplies && bridge.ZMQStream == nil {
		log.Println("ZMQStream is not initialized. Cannot handle command.")
		return
	}

	switch cmd {
	case commands.Url:
		log.Printf("Received URL command: %s", cmd)
		if !sendReplies {
			return
		}
		if err := bridge.ZMQStream.SendFrame([]byte(bridge.WebUrl), goczmq.FlagNone); err != nil {
			log.Printf("Failed to send URL response: %v", err)
		}
		return
	case commands.Wait:
		log.Printf("Received Wait command: %s", cmd)
		if !sendReplies {
			return
		}
		bridge.waitForWebsockets()
		return
	case commands.SetTarget:
		log.Printf("Received SetTarget command: %s", cmd)
		if len(frames) != 3 {
			if !sendReplies {
				return
			}
			bridge.sendExpected3FramesError(cmd)
			return
		}
		bridge.forwardToWebsockets(frames[2])
		if !sendReplies {
			return
		}
		if err := bridge.ZMQStream.SendFrame([]byte("ok"), goczmq.FlagNone); err != nil {
			log.Printf("Failed to send ok response for SetTarget command: %v", err)
		}
		return
	case commands.CaptureImage:
		log.Printf("Received CaptureImage command: %s", cmd)
		if !sendReplies {
			return
		}

		if len(frames) != 3 {
			bridge.sendExpected3FramesError(cmd)
			return
		}

		for !bridge.hasWebsocketConn() {
			select {
			case <-bridge.stopCh:
				return
			case <-time.After(100 * time.Millisecond):
			}
		}

		respCh := make(chan []byte, 1)
		bridge.setCaptureResponseChannel(respCh)
		defer bridge.setCaptureResponseChannel(nil)

		bridge.forwardToWebsockets(frames[2])

		select {
		case img := <-respCh:
			if err := bridge.ZMQStream.SendFrame(img, goczmq.FlagNone); err != nil {
				log.Printf("Failed to send capture_image payload: %v", err)
			}
		case <-bridge.stopCh:
			return
		}
		return
	case commands.GetScene:
		log.Printf("Received GetScene command: %s", cmd)
		if !sendReplies {
			return
		}
		var drawingCmds strings.Builder
		bridge.SceneTree.Walk(func(node *scene.TreeNode) {
			if node.Object != nil {
				if b, ok := node.Object.([]byte); ok {
					drawingCmds.WriteString(zmqCreateCommand(b))
				}
			}
			for _, p := range node.Properties {
				if b, ok := p.([]byte); ok {
					drawingCmds.WriteString(zmqCreateCommand(b))
				}
			}
			if node.Transform != nil {
				if b, ok := node.Transform.([]byte); ok {
					drawingCmds.WriteString(zmqCreateCommand(b))
				}
			}
			if node.Animation != nil {
				if b, ok := node.Animation.([]byte); ok {
					drawingCmds.WriteString(zmqCreateCommand(b))
				}
			}
		})
		jsBytes, err := meshcatgo.ViewerAssets.ReadFile("viewer_assets/dist/main.min.js")
		if err != nil {
			log.Printf("get_scene: failed to read main.min.js: %v", err)
			return
		}
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
	<head><meta charset=utf-8><title>MeshCat</title></head>
	<body>
		<div id="meshcat-pane"></div>
		<script>%s</script>
		<script>
			var viewer = new MeshCat.Viewer(document.getElementById("meshcat-pane"));
			%s
		</script>
		<style>
			body { margin: 0; }
			#meshcat-pane { width: 100vw; height: 100vh; overflow: hidden; }
		</style>
		<script id="embedded-json"></script>
	</body>
</html>`, string(jsBytes), drawingCmds.String())
		if err := bridge.ZMQStream.SendFrame([]byte(html), goczmq.FlagNone); err != nil {
			log.Printf("get_scene: failed to send HTML: %v", err)
		}
		return
	}

	// - Meshcat-specific commands (e.g. "set_transform", etc.)

	// Check to see if the command is a Meshcat-specific command
	if commands.IsMeshcatCommand(cmd) {
		if len(frames) != 3 {
			if !sendReplies {
				return
			}
			bridge.sendExpected3FramesError(cmd)
			return
		}

		// If so, then we need to extract the following important data:
		// - path to the relevant object,
		path := make([]string, 0)

		for _, part := range strings.Split(string(frames[1]), "/") {
			if part != "" {
				path = append(path, part)
			}
		}

		// - data payload (e.g. transform data, object data, etc.)
		data := frames[2]

		cacheHit := false

		switch cmd {
		case commands.SetTransform:
			log.Println("Received SetTransform command")

			// Get the node corresponding to the current path
			targetNode := bridge.SceneTree.GetPath(path)
			if targetNode == nil {
				log.Printf("Failed to find target node for path: %v", path)
				return
			}

			// Parse the transform data from the frames and apply it to the target node.
			targetNode.Transform = data // Placeholder: In reality, we would parse the transform data and set it appropriately. (Maybe this isn't necessary if we just want to store the raw data and let the client handle it?)
			bridge.forwardToWebsockets(data)

		case commands.SetObject:
			log.Println("Received SetObject command")

			// Get the node corresponding to the current path
			targetNode := bridge.SceneTree.GetPath(path)
			if targetNode == nil {
				log.Printf("Failed to find target node for path: %v", path)
				return
			}

			if existing, ok := targetNode.Object.([]byte); ok && bytes.Equal(existing, data) {
				cacheHit = true
			}

			// Set the object data on the target node AND clear the object's properties (since setting a new object should clear any existing properties).
			targetNode.Object = data
			targetNode.Properties = make([]any, 0)

			if !cacheHit {
				bridge.forwardToWebsockets(data)
			}

		case commands.SetProperty:
			log.Println("Received SetProperty command")

			// Get the node corresponding to the current path
			targetNode := bridge.SceneTree.GetPath(path)
			if targetNode == nil {
				log.Printf("Failed to find target node for path: %v", path)
				return
			}

			// Add the new property to the target node's properties.
			targetNode.Properties = append(targetNode.Properties, data)
			bridge.forwardToWebsockets(data)

		case commands.SetAnimation:
			log.Println("Received SetAnimation command")

			// Get the node corresponding to the current path
			targetNode := bridge.SceneTree.GetPath(path)
			if targetNode == nil {
				log.Printf("Failed to find target node for path: %v", path)
				return
			}

			// Set the animation data on the target node.
			targetNode.Animation = data
			bridge.forwardToWebsockets(data)

		case commands.Delete:
			log.Println("Received Delete command")

			// Check to see if the path has nonzero length (if not, then we are trying to delete the root node)
			if len(path) > 0 {
				parentPath := path[:len(path)-1]
				childKey := path[len(path)-1]

				// Get the parent node corresponding to the parent path
				parentNode := bridge.SceneTree.GetPath(parentPath)
				if parentNode == nil {
					log.Printf("Failed to find parent node for path: %v", parentPath)
					return
				}

				// Delete the child node from the parent's children map (if it exists)
				if parentNode.Children != nil {
					delete(parentNode.Children, childKey)
				}
			} else {
				// If the path is empty, then we are trying to delete the root node. We can interpret this as clearing the entire scene tree.
				bridge.SceneTree = scene.NewTreeNode()
			}
			bridge.forwardToWebsockets(data)

		}

		// For all meshcat commands, send an "ok" response
		// via the ZMQ stream to acknowledge that we received and processed the command.
		if sendReplies {
			if err := bridge.ZMQStream.SendFrame([]byte("ok"), goczmq.FlagNone); err != nil {
				log.Printf("failed to send ok response for command %s: %v", cmd, err)
			}
		}

		return
	}

	// Otherwise, we can log the unrecognized command and also send a response to the ZMQ stream.
	log.Printf("Received unrecognized command: %s", cmd)
	if !sendReplies {
		return
	}
	if err := bridge.ZMQStream.SendFrame([]byte("error: unrecognized command"), goczmq.FlagNone); err != nil {
		log.Printf("failed to send error response for command %s: %v", cmd, err)
	}
}

// WebsocketHandler is the handler for the websocket endpoint defined in the Gin router.
// It upgrades the HTTP connection to a websocket connection and then listens for messages from the client.
// When a message is received, it is... TBD
func (bridge *ZeroMQWebsocketBridge) WebsocketHandler(ctx *gin.Context) {
	// Collect the Gorilla websocket Upgrader
	// TODO(Kwesi): Describe what the upgrader is doing here?
	upgrader := bridge.GorillaUpgrader
	if upgrader == nil {
		upgrader = &websocket.Upgrader{}
	}

	w, r := ctx.Writer, ctx.Request
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	bridge.addWebsocketConn(c)
	defer func() {
		bridge.removeWebsocketConn(c)
		_ = c.Close()
	}()

	// Immediately send the current scene state so new clients render the
	// existing world before incremental updates arrive.
	bridge.sendScene(c)

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		if mt == websocket.TextMessage {
			imgBytes, err := decodeCaptureImagePayload(message)
			if err != nil {
				log.Printf("failed to parse websocket message: %v", err)
				continue
			}

			respCh := bridge.getCaptureResponseChannel()
			if respCh == nil {
				continue
			}

			select {
			case respCh <- imgBytes:
			default:
			}
		}
	}
}

// NewZeroMQWebsocketBridge creates a bridge configured with the default
// Meshcat web server port.
func NewZeroMQWebsocketBridge() *ZeroMQWebsocketBridge {
	return NewZeroMQWebsocketBridgeWithWebPort(DEFAULT_FILESERVER_PORT)
}

// NewZeroMQWebsocketBridgeWithWebPort creates a bridge and sets the advertised
// WebUrl using the provided HTTP port.
//
// The bridge still auto-selects an available ZMQ REP port independently.
func NewZeroMQWebsocketBridgeWithWebPort(webPort int) *ZeroMQWebsocketBridge {
	// Select a URL for the zmq server
	defaultPort := 6000

	bridge := ZeroMQWebsocketBridge{
		Host:      ZMQ_DEFAULT_HOST,
		SceneTree: scene.NewTreeNode(),
		wsPool:    make(map[*websocket.Conn]struct{}),
		stopCh:    make(chan struct{}),
	}

	_, zmqStream, foundPort, err := FindAvailablePort(bridge.SetUpZMQ, defaultPort, 100)
	if err != nil {
		log.Fatalf("Failed to set up ZMQ sockets: %v", err)
	}
	bridge.ZMQStream = zmqStream
	bridge.Port = foundPort
	bridge.ZMQUrl = GenerateZMQUrl("tcp", bridge.Host, foundPort)
	bridge.WebUrl = bridge.BuildWebUrl(webPort)

	// Report the zmq url and the web url
	log.Printf("ZeroMQ Websocket Bridge started at %s:%d", bridge.Host, bridge.Port)
	log.Printf("zmq_url: %s", bridge.ZMQUrl)
	log.Printf("web_url: %s", bridge.WebUrl)

	// Create the fileserver port

	return &bridge
}

func (bridge *ZeroMQWebsocketBridge) Stop() {
	bridge.stopOnce.Do(func() {
		if bridge.stopCh != nil {
			close(bridge.stopCh)
		}
	})
}

func (bridge *ZeroMQWebsocketBridge) Run() {
	// Starts running the ZeroMQ Websocket Bridge.
	//
	// TODO: Consider replacing the polling loop below with a goczmq.NewChanneler-based approach,
	// which exposes RecvC/SendC channels and allows a clean select{} over stopCh and incoming
	// ZMQ frames without busy-waiting or SetRcvtimeo.
	// Trade-offs to be aware of before doing so:
	//   - NewChanneler takes ownership of the socket; SetRcvtimeo and direct SendFrame calls
	//     from outside the channeler goroutine become unsafe (ZMQ sockets are not thread-safe).
	//   - All reply sends in HandleZMQFrameSlice (ZMQStream.SendFrame) must be routed through
	//     channeler.SendC instead, requiring changes at every call site.
	//   - The channeler's internal goroutine swallows errors silently; the current loop surfaces
	//     RecvMessageNoWait errors explicitly.
	//   - Calling channeler.Destroy() also destroys the underlying socket, so bridge.Destroy()
	//     must not be called afterward.
	if bridge.ZMQStream == nil {
		log.Println("ZMQ stream is nil; cannot start bridge loop")
		return
	}
	if bridge.stopCh == nil {
		bridge.stopCh = make(chan struct{})
	}

	// Keep reads responsive if we ever switch to blocking recv calls.
	bridge.ZMQStream.SetRcvtimeo(100)

	for {
		select {
		case <-bridge.stopCh:
			return
		default:
		}

		if !bridge.ZMQStream.Pollin() {
			// Avoid a busy-loop while waiting for incoming commands.
			time.Sleep(5 * time.Millisecond)
			continue
		}

		frames, err := bridge.ZMQStream.RecvMessageNoWait()
		if err != nil {
			// Poll readiness can race with recv availability.
			continue
		}

		bridge.HandleZMQFrameSlice(frames)
	}
}

func (bridge *ZeroMQWebsocketBridge) SetUpZMQ(port int) (*goczmq.Sock, *goczmq.Sock, error) {
	// Setup/Defaults
	defaultZMQMethod := "tcp"

	// Create a REP socket to match the Python meshcat server's zmq.REP socket.
	// REP enforces a strict receive-then-send cycle, which is the protocol used
	// by all meshcat clients (Python, Go, etc.).
	targetURL := GenerateZMQUrl(defaultZMQMethod, bridge.Host, port)
	log.Printf("Attempting to bind ZMQ REP socket to URL: %s", targetURL)
	repSocket := goczmq.NewSock(goczmq.Rep)
	if _, err := repSocket.Bind(targetURL); err != nil {
		log.Printf("Failed to bind ZMQ REP socket to URL %s: %v", targetURL, err)
		repSocket.Destroy()
		return nil, nil, err
	}

	log.Printf("Successfully bound ZMQ REP socket to URL: %s", targetURL)
	return nil, repSocket, nil
}

// zmqCreateCommand converts a raw msgpack blob into a JavaScript statement
// that the MeshCat viewer can execute via handle_command_bytearray.
// Base64 encoding is used to keep the output compact for large meshes.
func zmqCreateCommand(data []byte) string {
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf(
		`viewer.handle_command_bytearray(Uint8Array.from(atob(%q), c => c.charCodeAt(0)));`+"\n",
		b64,
	)
}

// BuildWebUrl returns the viewer URL for the given HTTP file-server port.
func (bridge *ZeroMQWebsocketBridge) BuildWebUrl(fileServerPort int) string {
	protocol := "http:"

	return fmt.Sprintf(
		"%s//%s:%d/static/",
		protocol,
		bridge.Host,
		fileServerPort,
	)
}
