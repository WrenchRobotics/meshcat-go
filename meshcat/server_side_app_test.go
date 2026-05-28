package meshcat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDefineWebsocketApp_RootRedirectsToStatic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverApp := NewMeshcatWebServerApplication()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	serverApp.WebRouter.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	if location != "/static/" {
		t.Fatalf("expected redirect location /static/, got %q", location)
	}
}

func TestDefineWebsocketApp_StaticIndexIsServedFromEmbeddedAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverApp := NewMeshcatWebServerApplication()
	req := httptest.NewRequest(http.MethodGet, "/static/", nil)
	resp := httptest.NewRecorder()

	serverApp.WebRouter.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	body := resp.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty response body for /static/")
	}
	if !strings.Contains(strings.ToLower(body), "<html") {
		t.Fatal("expected HTML content in /static/ response")
	}
}
