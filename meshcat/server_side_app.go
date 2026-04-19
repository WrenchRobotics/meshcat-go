package meshcat

import (
	"io/fs"
	"net/http"
	"strconv"

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
}

func (app *MeshcatWebServerApplication) RunWebServer(port int) {
	err := app.WebRouter.Run(":" + strconv.Itoa(port))
	if err != nil {
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
	go app.RunWebServer(8080)
}

func NewMeshcatWebServerApplication() MeshcatWebServerApplication {
	// Create a new zmq websocket bridge
	bridge := zmqserver.NewZeroMQWebsocketBridge()

	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Serve the JavaScript MeshCat viewer bundled as a git submodule.
	distFS, _ := fs.Sub(meshcatgo.ViewerAssets, "third_party/meshcat-js/dist")
	router.StaticFS("/viewer", http.FS(distFS))
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/viewer")
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
	}
}
