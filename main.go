package main

import (
	"embed"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/plinehan/www/sess"
)

//go:embed index.html template.html
var contentFS embed.FS

type pageData struct {
	BgColor   string
	TextColor string
	NextURL   string
	Image     string
}

type appServer struct {
	indexTmpl *template.Template
	pageTmpl  *template.Template
}

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{}),
	))
}

func newServer() (*appServer, error) {
	indexBytes, err := contentFS.ReadFile("index.html")
	if err != nil {
		return nil, err
	}
	pageBytes, err := contentFS.ReadFile("template.html")
	if err != nil {
		return nil, err
	}

	indexTmpl, err := template.New("index").Parse(string(indexBytes))
	if err != nil {
		return nil, err
	}
	pageTmpl, err := template.New("page").Parse(string(pageBytes))
	if err != nil {
		return nil, err
	}

	return &appServer{
		indexTmpl: indexTmpl,
		pageTmpl:  pageTmpl,
	}, nil
}

func methodGETOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

func (s *appServer) renderPage(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.pageTmpl.Execute(w, data); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s *appServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.indexTmpl.Execute(w, nil); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "duration", time.Since(start))
	})
}

func (s *appServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	mux.HandleFunc("/", methodGETOnly(s.handleIndex))
	mux.HandleFunc("/ofcourse", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderPage(w, pageData{
			BgColor:   "#FFFFFF",
			TextColor: "#000000",
			NextURL:   "/",
			Image:     "ofcourse.jpg",
		})
	}))
	mux.HandleFunc("/funnyman", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderPage(w, pageData{
			BgColor:   "#000000",
			TextColor: "#FFFFFF",
			NextURL:   "brown",
			Image:     "funnyman.jpg",
		})
	}))
	mux.HandleFunc("/brown", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderPage(w, pageData{
			BgColor:   "#000000",
			TextColor: "#FFFFFF",
			NextURL:   "nurbs",
			Image:     "brown.jpg",
		})
	}))
	mux.HandleFunc("/nurbs", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderPage(w, pageData{
			BgColor:   "#FFFFFF",
			TextColor: "#000000",
			NextURL:   "thenextlevel",
			Image:     "nurbs.jpg",
		})
	}))
	mux.HandleFunc("/thenextlevel", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderPage(w, pageData{
			BgColor:   "#FFFFFF",
			TextColor: "#000000",
			NextURL:   "dog",
			Image:     "thenextlevel.jpg",
		})
	}))
	mux.HandleFunc("/dog", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderPage(w, pageData{
			BgColor:   "#000000",
			TextColor: "#FFFFFF",
			NextURL:   "http://johnniemanzari.com",
			Image:     "dog.jpg",
		})
	}))

	mux.HandleFunc("/sudo/login", methodGETOnly(handleLogin))
	mux.HandleFunc("/sudo/callback", methodGETOnly(handleCallback))

	return withRequestLogging(mux)
}

func main() {
	_ = sess.SessionSigningKey()
	server, err := newServer()
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Warn("listening", "port", port, "url", "http://localhost:"+port)
	if err := http.ListenAndServe(":"+port, server.routes()); err != nil {
		log.Fatal(err)
	}
}
