package router

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/uzhinskiy/extractor/modules/front"
	"github.com/uzhinskiy/lib.go/helpers"
)

var (
	version = "extractor/0.0.1"
)

type apiRequest struct {
	Action string                 `json:"action,omitempty"` // Имя вызываемого метода*
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
	remoteIP := helpers.GetIP(r.RemoteAddr, r.Header.Get("X-Real-IP"), r.Header.Get("X-Forwarded-For"))

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Server", version)

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

	log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", 200, "\t", r.UserAgent())

	switch request.Action {
	case "get_repositories":
		{
			w.Write([]byte("{\"OK\"}"))
		}
	case "get_nodes":
		{
			w.Write([]byte("{\"OK\"}"))
		}

	case "get_snapshots":
		{
			w.Write([]byte("{\"OK\"}"))
		}

	case "get_snapshot":
		{
			w.Write([]byte("{\"OK\"}"))
		}

	case "restore":
		{
			w.Write([]byte("{\"OK\"}"))
		}

	default:
		{
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", "Invalid request method ", "\t", r.UserAgent())
			return

		}

	}

}

func doGet(url string, request apiRequest) (*RESPONSE_JSON, error) {
	serviceResp := new(RESPONSE_JSON)
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(helpers.Atoi(appConfig["netdialtimeout"])) * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var netClient = &http.Client{
		Timeout:   time.Second * time.Duration(helpers.Atoi(appConfig["netclienttimeout"])),
		Transport: netTransport,
	}

	toBackend, _ := json.Marshal(request)

	actionRequest, _ := http.NewRequest("POST", url, bytes.NewReader(toBackend))
	actionRequest.Header.Set("Content-Type", "application/json")
	actionRequest.Header.Set("Connection", "keep-alive")

	actionResult, err := netClient.Do(actionRequest)
	if actionResult != nil {
		defer actionResult.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	// response validation
	err = json.NewDecoder(actionResult.Body).Decode(&serviceResp)
	if err != nil {
		return nil, err
	}

	if serviceResp.IsEmpty() {
		return nil, errors.New("Empty response from backend")
	}
	return serviceResp, nil
}

func doPost(url string, request apiRequest) (*RESPONSE_JSON, error) {
	serviceResp := new(RESPONSE_JSON)
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(helpers.Atoi(appConfig["netdialtimeout"])) * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var netClient = &http.Client{
		Timeout:   time.Second * time.Duration(helpers.Atoi(appConfig["netclienttimeout"])),
		Transport: netTransport,
	}

	toBackend, _ := json.Marshal(request)

	actionRequest, _ := http.NewRequest("POST", url, bytes.NewReader(toBackend))
	actionRequest.Header.Set("Content-Type", "application/json")
	actionRequest.Header.Set("Connection", "keep-alive")

	actionResult, err := netClient.Do(actionRequest)
	if actionResult != nil {
		defer actionResult.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	// response validation
	err = json.NewDecoder(actionResult.Body).Decode(&serviceResp)
	if err != nil {
		return nil, err
	}

	if serviceResp.IsEmpty() {
		return nil, errors.New("Empty response from backend")
	}
	return serviceResp, nil
}
