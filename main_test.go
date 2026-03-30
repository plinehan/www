package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	s, err := newServer()
	if err != nil {
		t.Fatalf("newServer() error = %v", err)
	}
	return s.routes()
}

func TestRoutesReturnOK(t *testing.T) {
	handler := newTestHandler(t)
	paths := []string{
		"/",
		"/ofcourse",
		"/funnyman",
		"/brown",
		"/nurbs",
		"/thenextlevel",
		"/dog",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status for %s = %d, want %d", path, rec.Code, http.StatusOK)
			}
		})
	}
}

func TestLegacyRouteContent(t *testing.T) {
	tests := []struct {
		path         string
		wantImageRef string
		wantHrefRef  string
	}{
		{path: "/ofcourse", wantImageRef: "/images/ofcourse.jpg", wantHrefRef: `href="/"`},
		{path: "/funnyman", wantImageRef: "/images/funnyman.jpg", wantHrefRef: `href="brown"`},
		{path: "/brown", wantImageRef: "/images/brown.jpg", wantHrefRef: `href="nurbs"`},
		{path: "/nurbs", wantImageRef: "/images/nurbs.jpg", wantHrefRef: `href="thenextlevel"`},
		{path: "/thenextlevel", wantImageRef: "/images/thenextlevel.jpg", wantHrefRef: `href="dog"`},
		{path: "/dog", wantImageRef: "/images/dog.jpg", wantHrefRef: `href="http://johnniemanzari.com"`},
	}

	handler := newTestHandler(t)
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			body := rec.Body.String()
			if !strings.Contains(body, tc.wantImageRef) {
				t.Fatalf("body for %s missing image ref %q", tc.path, tc.wantImageRef)
			}
			if !strings.Contains(body, tc.wantHrefRef) {
				t.Fatalf("body for %s missing href %q", tc.path, tc.wantHrefRef)
			}
		})
	}
}

func TestImagesPathIsNotHandledByGetOnlyMiddleware(t *testing.T) {
	handler := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/images/does-not-exist.jpg", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, did not expect method not allowed for /images/*", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	handler := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/ofcourse", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow header = %q, want %q", got, http.MethodGet)
	}
}
