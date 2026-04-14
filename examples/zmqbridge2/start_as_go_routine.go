package main

import "github.com/WrenchRobotics/meshcat-go/servers/zmqserver"

func main() {
	// Create ZMQServer Websocket bridge and start it
	bridge := zmqserver.NewZeroMQWebsocketBridge()
	go bridge.Run() // Start the bridge in a separate goroutine

	// Create the web application that will pass information to
	// the bridge
	router := zmqserver.DefineWebsocketApp(bridge)

	// Start the web application
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}

}
