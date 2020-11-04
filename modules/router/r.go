// Copyright © 2020 Uzhinskiy Boris
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"crypto/tls"
	"errors"
	"io/ioutil"

	"bytes"
	"net"

	"time"

	"github.com/uzhinskiy/extractor/modules/config"
	"github.com/uzhinskiy/extractor/modules/front"
	"github.com/uzhinskiy/lib.go/helpers"
)

var (
	version = "extractor/0.0.5"
)

type Router struct {
	conf config.Config
}

type apiRequest struct {
	Action string                 `json:"action,omitempty"` // Имя вызываемого метода*
	Values map[string]interface{} `json:"values,omitempty"`
}

func Run(cnf config.Config) {
	rt := Router{}
	rt.conf = cnf
	http.HandleFunc("/", rt.FrontHandler)
	http.HandleFunc("/api/", rt.ApiHandler)
	http.ListenAndServe(":"+cnf.App.Port, nil)
}

// web-ui
func (rt *Router) FrontHandler(w http.ResponseWriter, r *http.Request) {
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

func (rt *Router) ApiHandler(w http.ResponseWriter, r *http.Request) {
	var request apiRequest
	var ok bool

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

	log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 200, "\t", r.UserAgent())

	switch request.Action {
	case "get_repositories":
		{
			response, err := rt.doGet(rt.conf.Elastic.Host + "_cat/repositories?format=json")
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			w.Write(response)
		}
	case "get_nodes":
		{
			response, err := rt.doGet(rt.conf.Elastic.Host + "_cat/nodes?format=json&h=ip,name,dt,du,dup,d")
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			w.Write(response)
		}

	case "get_snapshots":
		{
			var repo string
			if repo, ok = request.Values["repo"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			response, err := rt.doGet(rt.conf.Elastic.Host + "_cat/snapshots/" + repo + "?format=json")
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			w.Write(response)
		}

	case "get_snapshot":
		{
			var repo string
			var snap string
			if repo, ok = request.Values["repo"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}

			if snap, ok = request.Values["snapshot"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}

			response, err := rt.doGet(rt.conf.Elastic.Host + "_snapshot/" + repo + "/" + snap)
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			w.Write(response)
		}

	case "restore":
		{

			var repo string
			var snap string
			var pattern string
			var replacement string
			if repo, ok = request.Values["repo"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}

			if snap, ok = request.Values["snapshot"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			if pattern, ok = request.Values["pattern"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			if replacement, ok = request.Values["replacement"].(string); !ok {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}

			req := map[string]interface{}{
				"ignore_unavailable":   false,
				"include_global_state": false,
				"include_aliases":      false,
				"rename_pattern":       pattern,
				"rename_replacement":   replacement,
				"indices":              request.Values["indices"],
			}

			response, err := rt.doPost(rt.conf.Elastic.Host+"_snapshot/"+repo+"/"+snap+"/_restore?wait_for_completion=false", req)
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			w.Write(response)
		}

	default:
		{
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", "Invalid request method ", "\t", r.UserAgent())
			return

		}

	}

}

func (rt *Router) doGet(url string) ([]byte, error) {

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(60) * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var netClient = &http.Client{
		Timeout:   time.Second * time.Duration(60),
		Transport: netTransport,
	}

	actionRequest, _ := http.NewRequest("GET", url, nil)
	actionRequest.Header.Set("Content-Type", "application/json")
	actionRequest.Header.Set("Connection", "keep-alive")

	actionResult, err := netClient.Do(actionRequest)
	if actionResult != nil {
		defer actionResult.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if actionResult.StatusCode != 200 {
		return nil, errors.New("Wrong response: " + actionResult.Status)
	}

	body, err := ioutil.ReadAll(actionResult.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (rt *Router) doPost(url string, request map[string]interface{}) ([]byte, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(60) * time.Second,
		}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var netClient = &http.Client{
		Timeout:   time.Second * time.Duration(60),
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

	if actionResult.StatusCode != 200 {
		return nil, errors.New("Wrong response: " + actionResult.Status)
	}

	body, err := ioutil.ReadAll(actionResult.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
