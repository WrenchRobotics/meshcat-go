package zmq

import (
	"encoding/base64"
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

type ZeroMQWebsocketBridge struct {
	GorillaUpgrader *websocket.Upgrader
	ZMQUrl          string
	WebUrl          string
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
}

func (bridge *ZeroMQWebsocketBridge) removeWebsocketConn(conn *websocket.Conn) {
	bridge.wsMu.Lock()
	defer bridge.wsMu.Unlock()

	if bridge.wsPool == nil {
		return
	}

	delete(bridge.wsPool, conn)
}

func (bridge *ZeroMQWebsocketBridge) hasWebsocketConn() bool {
	bridge.wsMu.RLock()
	defer bridge.wsMu.RUnlock()

	return len(bridge.wsPool) > 0
}

func (bridge *ZeroMQWebsocketBridge) forwardToWebsockets(data []byte) {
	bridge.wsMu.RLock()
	defer bridge.wsMu.RUnlock()

	for conn := range bridge.wsPool {
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			log.Printf("failed to forward message to websocket: %v", err)
		}
	}
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

// HandleZMQFrameSlice is a helper function that takes in a slice of ZMQ frames and
// handles the input frames according to Meshcat's ZMQ protocol.
// It returns a string response that can be sent back to the client.
func (bridge *ZeroMQWebsocketBridge) HandleZMQFrameSlice(frames [][]byte) {
	// Input Checking
	if len(frames) == 0 {
		log.Println("Received empty frame slice")
		panic("Received empty frame slice")
	}

	// Attempt to handle the frames according to:
	// - Standard ZMQ commands (e.g. "url", "wait", etc.)
	cmd := string(frames[0])
	switch cmd {
	case commands.Url:
		log.Printf("Received URL command: %s", cmd)
		bridge.ZMQStream.SendFrame([]byte(bridge.WebUrl), goczmq.FlagNone)
		return
	case commands.Wait:
		log.Printf("Received Wait command: %s", cmd)
		bridge.waitForWebsockets()
		return
	case commands.SetTarget:
		log.Printf("Received SetTarget command: %s", cmd)
		bridge.forwardToWebsockets(frames[2])
		if err := bridge.ZMQStream.SendFrame([]byte("ok"), goczmq.FlagNone); err != nil {
			log.Printf("failed to send ok for set_target: %v", err)
		}
		return
	case commands.GetScene:
		log.Printf("Received GetScene command: %s", cmd)
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
		jsBytes, err := meshcatgo.ViewerAssets.ReadFile("third_party/meshcat-js/dist/main.min.js")
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

		case commands.SetObject:
			log.Println("Received SetObject command")

			// Get the node corresponding to the current path
			targetNode := bridge.SceneTree.GetPath(path)
			if targetNode == nil {
				log.Printf("Failed to find target node for path: %v", path)
				return
			}

			// Set the object data on the target node AND clear the object's properties (since setting a new object should clear any existing properties).
			targetNode.Object = data
			targetNode.Properties = make([]any, 0)

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

		}

		return
	}

	// Otherwise, we can log the unrecognized command and ignore it for now.
	log.Printf("Received unrecognized command: %s", cmd)

}

// WebsocketHandler is the handler for the websocket endpoint defined in the Gin router.
// It upgrades the HTTP connection to a websocket connection and then listens for messages from the client.
// When a message is received, it is... TBD
func (bridge *ZeroMQWebsocketBridge) WebsocketHandler(ctx *gin.Context) {
	// Collect the Gorilla websocket Upgrader
	// TODO(Kwesi): Describe what the upgrader is doing here?
	upgrader := bridge.GorillaUpgrader

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

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv:%s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func NewZeroMQWebsocketBridge() *ZeroMQWebsocketBridge {
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
	bridge.WebUrl = bridge.BuildWebUrl(DEFAULT_FILESERVER_PORT)

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
	repSocket := goczmq.NewSock(goczmq.Rep)
	if _, err := repSocket.Bind(targetURL); err != nil {
		repSocket.Destroy()
		return nil, nil, err
	}

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

func (bridge *ZeroMQWebsocketBridge) BuildWebUrl(fileServerPort int) string {
	protocol := "http:"

	return fmt.Sprintf(
		"%s//%s:%d/static/",
		protocol,
		bridge.Host,
		fileServerPort,
	)
}
