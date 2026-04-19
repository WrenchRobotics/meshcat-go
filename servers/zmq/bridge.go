package zmq

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/WrenchRobotics/meshcat-go/commands"
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
	stopCh          chan struct{}
	stopOnce        sync.Once
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
		stopCh:          make(chan struct{}),
	}
}

func (bridge *ZeroMQWebsocketBridge) Destroy() {
	if bridge.ZMQStream != nil {
		bridge.ZMQStream.Destroy()
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

	// Meshcat Protocol logic
	cmd := string(frames[0])
	switch cmd {
	case commands.Url:
		log.Printf("Received URL command: %s", cmd)
		bridge.ZMQStream.SendFrame([]byte(bridge.WebUrl), goczmq.FlagNone)

	// TODO(Kwesi): Implement the rest of the Meshcat ZMQ protocol here.
	default:
		log.Printf("Received unrecognized command: %s", cmd)
	}

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
	defer func() { _ = c.Close() }()
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
		Host: ZMQ_DEFAULT_HOST,
		stopCh: make(chan struct{}),
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

	// Create the stream socket to connect to the router
	targetURL := GenerateZMQUrl(defaultZMQMethod, bridge.Host, port)
	zmqStream, err := goczmq.NewStream(targetURL)
	if err != nil {
		return nil, nil, err
	}

	return nil, zmqStream, nil
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
