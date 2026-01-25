package main

import (
	"fmt"
	"log"

	"github.com/WrenchRobotics/meshcat-go/commands"
	arg "github.com/alexflint/go-arg"
	"github.com/gin-gonic/gin"
	"github.com/zeromq/goczmq"
)

type ZMQServerArguments struct {
	Port int `arg:"--port" help:"The port to bind to"`
}

func CreateApp() *gin.Engine {
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Redirect people from the main endpoint to the proper website.
	// router.GET("/ws/:roomId", func(c *gin.Context) {
	// 	roomId := c.Param("roomId")
	// 	chat.ServeWS(c, roomId, hub)
	// })

	return router
}

const DefaultZMQMethod string = "tcp"

func DefineLocalUrl(method string, host string, port int) string {
	return fmt.Sprintf(
		"%s://%s:%d",
		method,
		host,
		port,
	)
}

const SocketType_REP int = 4
const DefaultZMQPort int = 6000

func HandleZMQFrames(frames [][]byte) {
	// Process the received frames
	log.Printf("Received %d frames", len(frames))
	for i, frame := range frames {
		log.Printf("Frame %d: %s", i, string(frame))
	}

	// Detect what the message is and respond accordingly
	command := string(frames[0])
	switch command {
	case commands.Url:
		log.Println("Received URL request")
		// Respond with the URL
		// Note: In a real implementation, you would send this back via the ZMQ socket
		log.Println("Responding with URL: http://localhost:8080")
	case commands.Wait:
		log.Println("Received wait request")
		// Simulate waiting
		log.Println("Waiting for 2 seconds...")
		// time.Sleep(2 * time.Second)
		log.Println("Wait complete")
	default:
		log.Printf("Unknown request: %v", frames[0])
	}
}

func main() {
	// This is just a placeholder file to make sure the examples/zmqserver
	// directory gets included in the module.
	var args ZMQServerArguments
	arg.MustParse(&args)

	// // Create App Using the ZMQServerArguments
	// router := CreateApp()

	// Setup ZMQ
	// zmqSocket := goczmq.NewSock(SocketType_REP)
	// tcpPort, err := zmqSocket.Bind(DefineLocalUrl(DefaultZMQMethod, "localhost", DefaultZMQPort))

	router, err := goczmq.NewRouter(fmt.Sprintf("tcp://*:%d", DefaultZMQPort))
	if err != nil {
		log.Fatal(err)
	}
	defer router.Destroy()

	targetEndpoint := DefineLocalUrl(DefaultZMQMethod, "127.0.0.1", 5555)
	fmt.Println("Binding ZeroMQ REP socket to ", targetEndpoint)
	zmqStream, err := goczmq.NewStream(targetEndpoint)
	if err != nil {
		panic(err)
	}
	defer zmqStream.Destroy()

	// // Start the server on the specified URL
	// router.Run(fmt.Sprintf(":%d", args.Port))

	// Create a router socket and bind it to port 5555.

	// Create a dealer socket and connect it to the router.
	dealer, err := goczmq.NewDealer(fmt.Sprintf("tcp://127.0.0.1:%d", DefaultZMQPort))
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()

	log.Println("dealer created and connected")

	// Send an initial  message from the dealer to the router.
	// In this case, the message is a single frame with the content "url".
	dataForFrame := "url"
	err = dealer.SendFrame([]byte(dataForFrame), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dealer sent '%s'", dataForFrame)

	// Receive the message. Here we call RecvMessage, which
	// will return the message as a slice of frames ([][]byte).
	// Since this is a router socket that support async
	// request / reply, the first frame of the message will
	// be the routing frame.
	request, err := router.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("router received '%s' from '%v'", request[1], request[0])
	HandleZMQFrames(request[1:])

	// Send a reply. First we send the routing frame, which
	// lets the dealer know which client to send the message.
	// The FlagMore flag tells the router there will be more
	// frames in this message.
	err = router.SendFrame(request[0], goczmq.FlagMore)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("router sent 'World'")

	// Next send the reply. The FlagNone flag tells the router
	// that this is the last frame of the message.
	err = router.SendFrame([]byte("World"), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	// Receive the reply.
	reply, err := dealer.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dealer received '%s'", string(reply[0]))
	HandleZMQFrames(reply)

}
