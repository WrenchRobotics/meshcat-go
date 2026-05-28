package zmq

import (
	"fmt"
	"net"
	"testing"
)

func TestFindAvailableTCPPortReturnsBindablePort(t *testing.T) {
	port, err := FindAvailableTCPPort()
	if err != nil {
		t.Fatalf("FindAvailableTCPPort returned error: %v", err)
	}
	if port <= 0 {
		t.Fatalf("expected a positive TCP port, got %d", port)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("expected port %d to be bindable, got error: %v", port, err)
	}
	_ = listener.Close()
}