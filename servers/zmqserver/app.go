package zmqserver

import (
	"github.com/gin-gonic/gin"
)

func DefineWebsocketApp(bridge *ZeroMQWebsocketBridge) *gin.Engine {
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Define main endpoint
	router.GET("/", bridge.home)

	// Redirect people from the main endpoint to the proper website.
	// router.GET("/ws/:roomId", func(c *gin.Context) {
	// 	roomId := c.Param("roomId")
	// 	chat.ServeWS(c, roomId, hub)
	// })

	return router
}
