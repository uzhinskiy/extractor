package router

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/uzhinskiy/extractor/modules/front"
)

var (
	version = "extractor/0.0.1"
)

type apiRequest struct {
	Action string                 `json:"version,omitempty"` // Имя вызываемого метода*
	Values map[string]interface{} `json:"values,omitempty"`
}

func Run() {
	http.HandleFunc("/", FrontHandler)
	http.HandleFunc("/api/", ApiHandler)
	http.ListenAndServe(":9400", nil)
}

// web-ui
func FrontHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	if file == "/" {
		file = "/index.html"
	}
	cFile := strings.Replace(file, "/", "", 1)
	data, err := front.Asset(cFile)
	if err != nil {
		log.Println(err)
	}

	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", r.UserAgent())
	/* отправить его клиенту */
	contentType := mime.TypeByExtension(path.Ext(cFile))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Server", version)
	w.Write(data)
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	var request apiRequest

	defer r.Body.Close()
	remoteIP := getIP(r.RemoteAddr, r.Header.Get("X-Real-IP"), r.Header.Get("X-Forwarded-For"))

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", "Invalid request method ", "\t", r.UserAgent())
		return
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
		return
	}
	log.Println(request)
}
