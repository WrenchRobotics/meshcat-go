package zmq

import (
	"fmt"
	"testing"
	"time"

	"github.com/WrenchRobotics/meshcat-go/commands"
	"github.com/zeromq/goczmq"
)

func TestRunRespondsToURLCommand(t *testing.T) {
	rep := goczmq.NewSock(goczmq.Rep)
	if rep == nil {
		t.Fatal("failed to create REP socket")
	}

	port, err := rep.Bind("tcp://127.0.0.1:*")
	if err != nil {
		rep.Destroy()
		t.Fatalf("failed to bind REP socket: %v", err)
	}

	req := goczmq.NewSock(goczmq.Req)
	if req == nil {
		rep.Destroy()
		t.Fatal("failed to create REQ socket")
	}

	if err := req.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port)); err != nil {
		req.Destroy()
		rep.Destroy()
		t.Fatalf("failed to connect REQ socket: %v", err)
	}

	expectedURL := "http://127.0.0.1:6003/static/"
	bridge := &ZeroMQWebsocketBridge{
		ZMQStream: rep,
		WebUrl:    expectedURL,
		stopCh:    make(chan struct{}),
	}

	runDone := make(chan struct{})
	go func() {
		bridge.Run()
		close(runDone)
	}()

	// Give the run loop a brief moment to start polling.
	time.Sleep(20 * time.Millisecond)

	if err := req.SendFrame([]byte(commands.Url), goczmq.FlagNone); err != nil {
		bridge.Stop()
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
		}
		req.Destroy()
		rep.Destroy()
		t.Fatalf("failed to send URL command frame: %v", err)
	}

	req.SetRcvtimeo(1000)
	reply, err := req.RecvMessage()
	if err != nil {
		bridge.Stop()
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
		}
		req.Destroy()
		rep.Destroy()
		t.Fatalf("failed to receive URL response frame: %v", err)
	}

	if len(reply) != 1 {
		bridge.Stop()
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
		}
		req.Destroy()
		rep.Destroy()
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}

	if string(reply[0]) != expectedURL {
		bridge.Stop()
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
		}
		req.Destroy()
		rep.Destroy()
		t.Fatalf("unexpected URL reply: got %q, want %q", string(reply[0]), expectedURL)
	}

	bridge.Stop()
	select {
	case <-runDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Run loop did not stop in time")
	}

	req.Destroy()
	rep.Destroy()
}
