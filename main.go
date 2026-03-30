package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed index.html template.html
var contentFS embed.FS

type legacyPageData struct {
	BgColor   string
	TextColor string
	NextURL   string
	Image     string
}

type appServer struct {
	indexTmpl  *template.Template
	legacyTmpl *template.Template
}

func newServer() (*appServer, error) {
	indexBytes, err := contentFS.ReadFile("index.html")
	if err != nil {
		return nil, err
	}
	legacyBytes, err := contentFS.ReadFile("template.html")
	if err != nil {
		return nil, err
	}

	indexTmpl, err := template.New("index").Parse(string(indexBytes))
	if err != nil {
		return nil, err
	}
	legacyTmpl, err := template.New("legacy").Parse(string(legacyBytes))
	if err != nil {
		return nil, err
	}

	return &appServer{
		indexTmpl:  indexTmpl,
		legacyTmpl: legacyTmpl,
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

func (s *appServer) renderLegacyPage(w http.ResponseWriter, data legacyPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.legacyTmpl.Execute(w, data); err != nil {
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

func (s *appServer) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", methodGETOnly(s.handleIndex))
	mux.HandleFunc("/ofcourse", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderLegacyPage(w, legacyPageData{
			BgColor:   "#FFFFFF",
			TextColor: "#000000",
			NextURL:   "/",
			Image:     "ofcourse.jpg",
		})
	}))
	mux.HandleFunc("/funnyman", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderLegacyPage(w, legacyPageData{
			BgColor:   "#000000",
			TextColor: "#FFFFFF",
			NextURL:   "brown",
			Image:     "funnyman.jpg",
		})
	}))
	mux.HandleFunc("/brown", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderLegacyPage(w, legacyPageData{
			BgColor:   "#000000",
			TextColor: "#FFFFFF",
			NextURL:   "nurbs",
			Image:     "brown.jpg",
		})
	}))
	mux.HandleFunc("/nurbs", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderLegacyPage(w, legacyPageData{
			BgColor:   "#FFFFFF",
			TextColor: "#000000",
			NextURL:   "thenextlevel",
			Image:     "nurbs.jpg",
		})
	}))
	mux.HandleFunc("/thenextlevel", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderLegacyPage(w, legacyPageData{
			BgColor:   "#FFFFFF",
			TextColor: "#000000",
			NextURL:   "dog",
			Image:     "thenextlevel.jpg",
		})
	}))
	mux.HandleFunc("/dog", methodGETOnly(func(w http.ResponseWriter, _ *http.Request) {
		s.renderLegacyPage(w, legacyPageData{
			BgColor:   "#000000",
			TextColor: "#FFFFFF",
			NextURL:   "http://johnniemanzari.com",
			Image:     "dog.jpg",
		})
	}))

	return mux
}

func main() {
	server, err := newServer()
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, server.routes()); err != nil {
		log.Fatal(err)
	}
}
