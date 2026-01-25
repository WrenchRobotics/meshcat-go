package zmqserver

import (
	"fmt"

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
