package meshcat

import (
	"context"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	meshcatgo "github.com/WrenchRobotics/meshcat-go"
	zmqserver "github.com/WrenchRobotics/meshcat-go/servers/zmq"
	"github.com/gin-gonic/gin"
)

// MeshcatWebServerApplication is a lightweight application that runs
// the web server that:
//
// 1. Serves the Meshcat viewer using the local javascript assets (embedded using Go's embed package),
//
// 2. Defines a websocket endpoint that can be used to pass information to the ZeroMQ bridge.
type MeshcatWebServerApplication struct {
	// Bridge is the ZeroMQWebsocketBridge that defines the websocket endpoint and handles communication with the ZeroMQ server.
	Bridge *zmqserver.ZeroMQWebsocketBridge

	// WebRouter is the Gin router that defines the web application.
	// It is defined as a field in the struct so that it can be used to start the web server in a separate goroutine.
	WebRouter *gin.Engine

	// WebPort is the port used by the HTTP server and advertised by the bridge web URL.
	WebPort int

	// WebServer serves the Gin router on either a reserved listener or a port-bound socket.
	WebServer *http.Server

	// Listener is the reserved HTTP listener used by the default constructor.
	Listener net.Listener
}

func (app *MeshcatWebServerApplication) RunWebServer() {
	if app.WebServer == nil {
		app.WebServer = &http.Server{Handler: app.WebRouter}
	}

	if app.Listener != nil {
		if err := app.WebServer.Serve(app.Listener); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
		return
	}

	if err := app.WebServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// Start is the method that starts both:
//
// 1. The ZeroMQWebsocketBridge that defines the websocket endpoint and handles communication with the ZeroMQ server.
//
// 2. The Gin web server that serves the Meshcat viewer and defines the websocket endpoint.
func (app *MeshcatWebServerApplication) Start() {
	// Start the ZeroMQWebsocketBridge in a separate goroutine
	go app.Bridge.Run()

	// Start the Gin web server in a separate goroutine
	go app.RunWebServer()
}

func (app *MeshcatWebServerApplication) Stop() error {
	app.Bridge.Stop()

	if app.WebServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return app.WebServer.Shutdown(ctx)
}

// NewMeshcatWebServerApplication creates a Meshcat web server application
// using an available HTTP port on the local machine.
func NewMeshcatWebServerApplication() MeshcatWebServerApplication {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		_ = listener.Close()
		panic("listener did not return a TCP address")
	}

	app := NewMeshcatWebServerApplicationWithPort(addr.Port)
	app.Listener = listener
	app.WebServer = &http.Server{Handler: app.WebRouter}

	return app
}

// NewMeshcatWebServerApplicationWithPort creates a Meshcat web server application
// using the provided HTTP port.
//
// The same webPort value is used to:
//  1. start the Gin HTTP server, and
//  2. build the bridge WebUrl advertised to ZMQ clients.
//
// Keeping those values aligned ensures the viewer opens on the same port that
// the browser can connect to for websocket updates.
func NewMeshcatWebServerApplicationWithPort(webPort int) MeshcatWebServerApplication {
	// Create a new zmq websocket bridge
	bridge := zmqserver.NewZeroMQWebsocketBridgeWithWebPort(webPort)

	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Serve the JavaScript MeshCat viewer bundled as a git submodule.
	distFS, err := fs.Sub(meshcatgo.ViewerAssets, "third_party/meshcat-js/dist")
	if err != nil {
		panic(err)
	}
	router.StaticFS("/static", http.FS(distFS))
	router.GET("/", func(c *gin.Context) {
		// Meshcat JS defaults to ws://<host> (no path), so accept websocket upgrades on root.
		if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			bridge.WebsocketHandler(c)
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/static/")
	})

	// Define websocket endpoint
	router.GET("/ws", bridge.WebsocketHandler)

	// Redirect people from the main endpoint to the proper website.
	// router.GET("/ws/:roomId", func(c *gin.Context) {
	// 	roomId := c.Param("roomId")
	// 	chat.ServeWS(c, roomId, hub)
	// })

	// Define a health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return MeshcatWebServerApplication{
		Bridge:    bridge,
		WebRouter: router,
		WebPort:   webPort,
		WebServer: &http.Server{
			Addr:    ":" + strconv.Itoa(webPort),
			Handler: router,
		},
	}
}
