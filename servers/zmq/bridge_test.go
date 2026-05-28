package zmq

import (
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WrenchRobotics/meshcat-go/commands"
	"github.com/WrenchRobotics/meshcat-go/scene"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

func TestDecodeCaptureImagePayload(t *testing.T) {
	want := []byte{0x89, 0x50, 0x4e, 0x47}
	msg := fmt.Sprintf(`{"data":"data:image/png;base64,%s"}`,
		base64.StdEncoding.EncodeToString(want),
	)

	got, err := decodeCaptureImagePayload([]byte(msg))
	if err != nil {
		t.Fatalf("decodeCaptureImagePayload returned error: %v", err)
	}

	if string(got) != string(want) {
		t.Fatalf("decoded bytes mismatch: got %v want %v", got, want)
	}
}

func TestDecodeCaptureImagePayloadRejectsInvalidJSON(t *testing.T) {
	if _, err := decodeCaptureImagePayload([]byte("not-json")); err == nil {
		t.Fatal("expected decodeCaptureImagePayload to fail for invalid JSON")
	}
}

func TestWebsocketConnectReceivesSceneSnapshot(t *testing.T) {
	bridge := &ZeroMQWebsocketBridge{
		SceneTree:       scene.NewTreeNode(),
		GorillaUpgrader: &websocket.Upgrader{},
		wsPool:          make(map[*websocket.Conn]struct{}),
	}

	objectBlob := []byte{0x01, 0x02, 0x03}
	transformBlob := []byte{0x04, 0x05, 0x06}
	node := bridge.SceneTree.GetPath([]string{"meshcat", "box"})
	node.Object = objectBlob
	node.Transform = transformBlob

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", bridge.WebsocketHandler)
	httpServer := httptest.NewServer(router)
	t.Cleanup(httpServer.Close)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket test server: %v", err)
	}
	t.Cleanup(func() { _ = wsConn.Close() })

	if err := wsConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set websocket read deadline: %v", err)
	}

	mt1, msg1, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read first snapshot message: %v", err)
	}
	if mt1 != websocket.BinaryMessage {
		t.Fatalf("unexpected first message type: got %d want %d", mt1, websocket.BinaryMessage)
	}

	mt2, msg2, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read second snapshot message: %v", err)
	}
	if mt2 != websocket.BinaryMessage {
		t.Fatalf("unexpected second message type: got %d want %d", mt2, websocket.BinaryMessage)
	}

	got := map[string]bool{
		string(msg1): true,
		string(msg2): true,
	}
	if !got[string(objectBlob)] || !got[string(transformBlob)] {
		t.Fatalf("snapshot did not include expected blobs; got messages %v and %v", msg1, msg2)
	}
}

func TestCaptureImageIntegrationViaWebsocket(t *testing.T) {
	bridge, req := setupRunningBridgeWithReqRep(t)
	bridge.GorillaUpgrader = &websocket.Upgrader{}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", bridge.WebsocketHandler)
	httpServer := httptest.NewServer(router)
	t.Cleanup(httpServer.Close)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket test server: %v", err)
	}
	t.Cleanup(func() { _ = wsConn.Close() })

	wantImage := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a}
	wantPayload := []byte("capture-cmd-payload")

	wsDone := make(chan error, 1)
	go func() {
		mt, payload, err := wsConn.ReadMessage()
		if err != nil {
			wsDone <- fmt.Errorf("websocket read failed: %w", err)
			return
		}
		if mt != websocket.BinaryMessage {
			wsDone <- fmt.Errorf("unexpected websocket message type: got %d want %d", mt, websocket.BinaryMessage)
			return
		}
		if string(payload) != string(wantPayload) {
			wsDone <- fmt.Errorf("unexpected websocket payload: got %q want %q", string(payload), string(wantPayload))
			return
		}

		msg := fmt.Sprintf(`{"data":"data:image/png;base64,%s"}`,
			base64.StdEncoding.EncodeToString(wantImage),
		)
		if err := wsConn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			wsDone <- fmt.Errorf("websocket write failed: %w", err)
			return
		}

		wsDone <- nil
	}()

	if err := req.SendFrame([]byte(commands.CaptureImage), goczmq.FlagMore); err != nil {
		t.Fatalf("failed to send capture_image cmd frame: %v", err)
	}
	if err := req.SendFrame([]byte(""), goczmq.FlagMore); err != nil {
		t.Fatalf("failed to send capture_image path frame: %v", err)
	}
	if err := req.SendFrame(wantPayload, goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send capture_image payload frame: %v", err)
	}

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive capture_image reply: %v", err)
	}
	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}
	if string(reply[0]) != string(wantImage) {
		t.Fatalf("unexpected image reply bytes: got %v want %v", reply[0], wantImage)
	}

	select {
	case err := <-wsDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for websocket exchange to finish")
	}
}

func TestSetTargetReturnsErrorWhenFramesMalformed(t *testing.T) {
	_, req := setupRunningBridgeWithReqRep(t)

	if err := req.SendFrame([]byte(commands.SetTarget), goczmq.FlagMore); err != nil {
		t.Fatalf("failed to send set_target cmd frame: %v", err)
	}
	if err := req.SendFrame([]byte("/meshcat"), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send set_target path frame: %v", err)
	}

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive set_target malformed response: %v", err)
	}
	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}
	if string(reply[0]) != "error: expected 3 frames" {
		t.Fatalf("unexpected malformed response: got %q", string(reply[0]))
	}
}

func TestSetObjectReturnsErrorWhenFramesMalformed(t *testing.T) {
	_, req := setupRunningBridgeWithReqRep(t)

	if err := req.SendFrame([]byte(commands.SetObject), goczmq.FlagMore); err != nil {
		t.Fatalf("failed to send set_object cmd frame: %v", err)
	}
	if err := req.SendFrame([]byte("/meshcat/box"), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send set_object path frame: %v", err)
	}

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive set_object malformed response: %v", err)
	}
	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}
	if string(reply[0]) != "error: expected 3 frames" {
		t.Fatalf("unexpected malformed response: got %q", string(reply[0]))
	}
}

func TestCaptureImageReturnsErrorWhenFramesMalformed(t *testing.T) {
	_, req := setupRunningBridgeWithReqRep(t)

	if err := req.SendFrame([]byte(commands.CaptureImage), goczmq.FlagMore); err != nil {
		t.Fatalf("failed to send capture_image cmd frame: %v", err)
	}
	if err := req.SendFrame([]byte(""), goczmq.FlagNone); err != nil {
		t.Fatalf("failed to send capture_image path frame: %v", err)
	}

	req.SetRcvtimeo(2000)
	reply, err := req.RecvMessage()
	if err != nil {
		t.Fatalf("failed to receive capture_image malformed response: %v", err)
	}
	if len(reply) != 1 {
		t.Fatalf("expected 1 reply frame, got %d", len(reply))
	}
	if string(reply[0]) != "error: expected 3 frames" {
		t.Fatalf("unexpected malformed response: got %q", string(reply[0]))
	}
}

func TestSetObjectCacheHitDoesNotForwardTwice(t *testing.T) {
	bridge, req := setupRunningBridgeWithReqRep(t)
	bridge.GorillaUpgrader = &websocket.Upgrader{}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", bridge.WebsocketHandler)
	httpServer := httptest.NewServer(router)
	t.Cleanup(httpServer.Close)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket test server: %v", err)
	}
	t.Cleanup(func() { _ = wsConn.Close() })

	path := []byte("/meshcat/box")
	objPayload := []byte("same-object-bytes")

	for i := 0; i < 2; i++ {
		if err := req.SendFrame([]byte(commands.SetObject), goczmq.FlagMore); err != nil {
			t.Fatalf("failed to send set_object cmd frame: %v", err)
		}
		if err := req.SendFrame(path, goczmq.FlagMore); err != nil {
			t.Fatalf("failed to send set_object path frame: %v", err)
		}
		if err := req.SendFrame(objPayload, goczmq.FlagNone); err != nil {
			t.Fatalf("failed to send set_object payload frame: %v", err)
		}

		req.SetRcvtimeo(2000)
		reply, err := req.RecvMessage()
		if err != nil {
			t.Fatalf("failed to receive set_object response: %v", err)
		}
		if len(reply) != 1 || string(reply[0]) != "ok" {
			t.Fatalf("unexpected set_object response: %v", reply)
		}
	}

	if err := wsConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed setting websocket read deadline: %v", err)
	}
	mt, msg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed reading first forwarded set_object payload: %v", err)
	}
	if mt != websocket.BinaryMessage {
		t.Fatalf("unexpected message type for first set_object: got %d want %d", mt, websocket.BinaryMessage)
	}
	if string(msg) != string(objPayload) {
		t.Fatalf("unexpected first set_object payload: got %q want %q", string(msg), string(objPayload))
	}

	if err := wsConn.SetReadDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
		t.Fatalf("failed setting websocket read deadline: %v", err)
	}
	if _, _, err := wsConn.ReadMessage(); err == nil {
		t.Fatal("expected second identical set_object to be suppressed by cache-hit logic")
	}
}
