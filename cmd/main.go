package main

import (
        "log"
        "mime"
        "net/http"
        "path"
        "strings"
	"github.com/uzhinskiy/extractor/modules/front"
)

func staticHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	if file == "/" {
		file = "/index.html"
	}
	cFile := strings.Replace(file, "/", "", 1)
  data, err := Asset(cFile)
	if err != nil {
		log.Println(err)
	}

	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", r.UserAgent())
	/* отправить его клиенту */
	contentType := mime.TypeByExtension(path.Ext(cFile))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Server", "gohttp/0.2")
	w.Write(data)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

  
}

func main() {
        http.HandleFunc("/", staticHandler)
        http.HandleFunc("/api/", apiHandler)
        http.ListenAndServe(":9400", nil)
}
