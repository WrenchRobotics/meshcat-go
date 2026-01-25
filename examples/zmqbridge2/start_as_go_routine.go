package main

import "github.com/WrenchRobotics/meshcat-go/servers/zmqserver"

func main() {
	zmqserver.StartZMQServerAsGoRoutine()
}
