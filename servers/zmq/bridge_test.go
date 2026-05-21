package zmq

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/WrenchRobotics/meshcat-go/commands"
	"github.com/WrenchRobotics/meshcat-go/scene"
	"github.com/zeromq/goczmq"
)

func setupRunningBridgeWithReqRep(t *testing.T) (*ZeroMQWebsocketBridge, *goczmq.Sock) {
	t.Helper()

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

	bridge := &ZeroMQWebsocketBridge{
		ZMQStream: rep,
		SceneTree: scene.NewTreeNode(),
		stopCh:    make(chan struct{}),
	}

	runDone := make(chan struct{})
	go func() {
		bridge.Run()
		close(runDone)
	}()

	// Give the run loop a brief moment to start polling.
	time.Sleep(20 * time.Millisecond)

	t.Cleanup(func() {
		bridge.Stop()
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
			t.Fatal("Run loop did not stop in time")
		}
		req.Destroy()
		rep.Destroy()
	})

	return bridge, req
}

// TestRunRespondsToURLCommand is a sanity check to make sure that
// the Run method of the ZeroMQWebsocketBridge correctly responds to
// a URL command with the expected Web URL.
//
// It should return the value of the WebUrl field of the bridge, which is supposed to be
// set before Run is called. In this case, we will set it to:
//
// http://127.0.0.1:6003/static/
func TestRunRespondsToURLCommand(t *testing.T) {
	// Create a REP socket to act as the ZMQStream for the bridge,
	// and a REQ socket to send commands to it.
	// - Rep socket will be created and bound to a random port on localhost.
	rep := goczmq.NewSock(goczmq.Rep)
	if rep == nil {
		t.Fatal("failed to create REP socket")
	}

	port, err := rep.Bind("tcp://127.0.0.1:*")
	if err != nil {
		rep.Destroy()
		t.Fatalf("failed to bind REP socket: %v", err)
	}

	// - Req socket will connect to the REP socket's endpoint.
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

func TestRunWaitRespondsImmediatelyWhenWebsocketExists(t *testing.T) {
	bridge, req := setupRunningBridgeWithReqRep(t)
	bridge.addWebsocketConn(nil)

	if err := req.SendFrame([]byte(commands.Wait), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send wait command frame: %v", err)
	}

	req.SetRcvtimeo(1000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive wait response frame: %v", err)
	}

	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}

	if string(reply[0]) != "ok" {
		t.Fatalf("unexpected wait reply: got %q, want %q", string(reply[0]), "ok")
	}
}

func TestRunWaitRespondsAfterWebsocketConnects(t *testing.T) {
	bridge, req := setupRunningBridgeWithReqRep(t)

	if err := req.SendFrame([]byte(commands.Wait), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send wait command frame: %v", err)
	}

	go func() {
		time.Sleep(150 * time.Millisecond)
		bridge.addWebsocketConn(nil)
	}()

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive wait response frame: %v", err)
	}

	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}

	if string(reply[0]) != "ok" {
		t.Fatalf("unexpected wait reply: got %q, want %q", string(reply[0]), "ok")
	}
}

// TestHandleZMQGetSceneReturnsHTML verifies that the get_scene command causes
// the bridge to send back a complete MeshCat HTML page over ZMQ.
// It sets up a scene tree with one object and one transform, then checks that
// the reply contains the viewer bootstrap code and a drawing command for each blob.
func TestHandleZMQGetSceneReturnsHTML(t *testing.T) {
	bridge, req := setupRunningBridgeWithReqRep(t)

	// Populate the scene tree with a node that has an object and a transform.
	objectBlob := []byte{0x01, 0x02, 0x03}
	transformBlob := []byte{0x04, 0x05, 0x06}
	node := bridge.SceneTree.GetPath([]string{"meshcat", "box"})
	node.Object = objectBlob
	node.Transform = transformBlob

	if err := req.SendFrame([]byte(commands.GetScene), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send get_scene frame: %v", err)
	}

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive get_scene reply: %v", err)
	}
	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}

	html := string(reply[0])

	for _, want := range []string{
		"<!DOCTYPE html>",
		"MeshCat.Viewer",
		"handle_command_bytearray",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("expected HTML to contain %q", want)
		}
	}
}

// TestHandleZMQGetSceneEmptyTree verifies that get_scene on an empty scene tree
// still returns a valid HTML page (just no drawing commands).
func TestHandleZMQGetSceneEmptyTree(t *testing.T) {
	_, req := setupRunningBridgeWithReqRep(t)

	if err := req.SendFrame([]byte(commands.GetScene), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send get_scene frame: %v", err)
	}

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive get_scene reply: %v", err)
	}
	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}

	html := string(reply[0])
	for _, want := range []string{
		"<!DOCTYPE html>",
		"MeshCat.Viewer",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("expected HTML to contain %q", want)
		}
	}
}

func TestRunWaitReturnsWithoutReplyWhenStopped(t *testing.T) {
	bridge, req := setupRunningBridgeWithReqRep(t)

	if err := req.SendFrame([]byte(commands.Wait), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send wait command frame: %v", err)
	}

	go func() {
		time.Sleep(150 * time.Millisecond)
		bridge.Stop()
	}()

	req.SetRcvtimeo(500)
	if _, err := req.RecvMessage(); err == nil {
		t.Fatal("expected wait recv to time out when bridge stops before websocket connects")
	}
}
