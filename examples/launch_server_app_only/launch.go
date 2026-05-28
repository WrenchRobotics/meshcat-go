package main

import (
	"github.com/WrenchRobotics/meshcat-go/meshcat"
)

func main() {
	// Create the web application that will pass information to
	// the bridge
	meshcatServerApp := meshcat.NewMeshcatWebServerApplication()

	// Start the web application and the bridge
	meshcatServerApp.Start()

	// Block the main thread indefinitely (since the web server and bridge are running in separate goroutines)
	select {}

}
