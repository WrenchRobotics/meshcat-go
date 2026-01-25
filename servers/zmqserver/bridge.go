package zmqserver

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zeromq/goczmq"
)

type ZeroMQWebsocketBridge struct {
	GorillaUpgrader *websocket.Upgrader
	ZMQUrl          string
	Host            string
	Port            int
	CertificateFile string
	KeyFile         string
	NGrokHttpTunnel bool
	ZMQStream       *goczmq.Sock
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
		NGrokHttpTunnel: ngrokHttpTunnel,
	}
}

func (bridge *ZeroMQWebsocketBridge) Destroy() {
	if bridge.ZMQStream != nil {
		bridge.ZMQStream.Destroy()
	}
}

func (bridge *ZeroMQWebsocketBridge) home(ctx *gin.Context) {
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
	}

	_, zmqStream, foundPort, err := FindAvailablePort(bridge.SetUpZMQ, defaultPort, 100)
	if err != nil {
		log.Fatalf("Failed to set up ZMQ sockets: %v", err)
	}
	bridge.ZMQStream = zmqStream
	bridge.Port = foundPort
	bridge.ZMQUrl = GenerateZMQUrl("tcp", bridge.Host, foundPort)

	// Report the zmq url and the web url
	log.Printf("ZeroMQ Websocket Bridge started at %s:%d", bridge.Host, bridge.Port)
	log.Printf("zmq_url: %s", bridge.ZMQUrl)
	log.Printf("web_url: %s", bridge.WebUrl(6003))

	// Create the fileserver port

	return &bridge
}

func (bridge *ZeroMQWebsocketBridge) Run() {
	// Starts running the ZeroMQ Websocket Bridge
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

func (bridge *ZeroMQWebsocketBridge) WebUrl(fileServerPort int) string {
	protocol := "http:"

	return fmt.Sprintf(
		"%s//%s:%d/static/",
		protocol,
		bridge.Host,
		fileServerPort,
	)
}
