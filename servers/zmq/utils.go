package zmq

import (
	"fmt"
	"net"

	"github.com/zeromq/goczmq"
)

func GenerateZMQUrl(zmqMethod string, host string, port int) string {
	return fmt.Sprintf(
		"%s://%s:%d",
		zmqMethod,
		host,
		port,
	)
}

func FindAvailablePort(
	setupFunction func(int) (*goczmq.Sock, *goczmq.Sock, error),
	defaultPort int,
	maxAttempts int,
) (*goczmq.Sock, *goczmq.Sock, int, error) {
	// Try to find an available port starting from defaultPort
	for port := defaultPort; port < defaultPort+maxAttempts; port++ {
		candidateZMQSocket, candidateZMQStream, err := setupFunction(port)
		if err == nil {
			return candidateZMQSocket, candidateZMQStream, port, err
		}
	}
	return nil, nil, -1, fmt.Errorf("failed to find an available port")
}

func FindAvailableTCPPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("listener did not return a TCP address")
	}

	return addr.Port, nil
}
