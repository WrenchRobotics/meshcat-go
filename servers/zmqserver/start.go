package zmqserver

import (
	"github.com/zeromq/goczmq"
)

func StartZMQServerAsGoRoutine() (*goczmq.Sock, string, string, error) {
	// Create ZMQServer Websocket bridge and start it
	bridge := NewZeroMQWebsocketBridge()

	// Start the bridge in a separate goroutine
	go bridge.Run()

	// Read from the log file to get the URL and return it
	return bridge.ZMQStream, bridge.Host, string(bridge.Port), nil
}
