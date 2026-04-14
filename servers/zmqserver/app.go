package zmqserver

import (
	"io/fs"
	"net/http"

	meshcatgo "github.com/WrenchRobotics/meshcat-go"
	"github.com/gin-gonic/gin"
)

type ZMQServerWebApplication struct {
	Bridge *ZeroMQWebsocketBridge
}

func DefineWebsocketApp(bridge *ZeroMQWebsocketBridge) *gin.Engine {
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Serve the JavaScript MeshCat viewer bundled as a git submodule.
	distFS, _ := fs.Sub(meshcatgo.ViewerAssets, "third_party/meshcat-js/dist")
	router.StaticFS("/viewer", http.FS(distFS))
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/viewer")
	})

	// Define websocket endpoint
	router.GET("/ws", bridge.home)

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

	return router
}
