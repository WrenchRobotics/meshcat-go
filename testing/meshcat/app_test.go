package zmqserver_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WrenchRobotics/meshcat-go/meshcat"
	"github.com/gin-gonic/gin"
)

func TestDefineWebsocketApp_RootRedirectsToViewer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverApp := meshcat.NewMeshcatWebServerApplication()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	serverApp.WebRouter.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	if location != "/viewer" {
		t.Fatalf("expected redirect location /viewer, got %q", location)
	}
}

func TestDefineWebsocketApp_ViewerIndexIsServedFromEmbeddedAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverApp := meshcat.NewMeshcatWebServerApplication()
	req := httptest.NewRequest(http.MethodGet, "/viewer/", nil)
	resp := httptest.NewRecorder()

	serverApp.WebRouter.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	body := resp.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty response body for /viewer/")
	}
	if !strings.Contains(strings.ToLower(body), "<html") {
		t.Fatal("expected HTML content in /viewer/ response")
	}
}
