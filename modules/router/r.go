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
	"fmt"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/uzhinskiy/extractor/modules/config"
	"github.com/uzhinskiy/extractor/modules/front"
	"github.com/uzhinskiy/extractor/modules/version"
	"github.com/uzhinskiy/lib.go/helpers"
)

type Router struct {
	conf  config.Config
	nc    *http.Client
	nodes nodesStatus
}

type apiRequest struct {
	Action string                 `json:"action,omitempty"` // Имя вызываемого метода*
	Values map[string]interface{} `json:"values,omitempty"`
}

type snapStatus struct {
	Snapshots []struct {
		Snapshot string `json:"snapshot,omitempty"`
		State    string `json:"state,omitempty"`
		Indices  map[string]struct {
			Stats struct {
				Total struct {
					Size int32 `json:"size_in_bytes,omitempty"`
				} `json:"total,omitempty"`
			} `json:"stats,omitempty"`
		} `json:"indices,omitempty"`
	} `json:"snapshots,omitempty"`
}

type singleNode struct {
	Ip       string `json:"ip,omitempty"`
	Name     string `json:"name,omitempty"`
	Dt       string `json:"dt,omitempty"`
	Du       string `json:"du,omitempty"`
	Dup      string `json:"dup,omitempty"`
	D        string `json:"d,omitempty"`
	DiskFree int
}

type nodesStatus struct {
	sync.RWMutex
	nlist []singleNode
}

func Run(cnf config.Config) {
	rt := Router{}
	rt.conf = cnf
	rt.netClientPrepare()
	_, err := rt.getNodes()
	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", rt.FrontHandler)
	http.HandleFunc("/api/", rt.ApiHandler)
	http.ListenAndServe(":"+cnf.App.Port, nil)
}

// web-ui
func (rt *Router) FrontHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	remoteIP := helpers.GetIP(r.RemoteAddr, r.Header.Get("X-Real-IP"), r.Header.Get("X-Forwarded-For"))
	if file == "/" {
		file = "/index.html"
	}
	cFile := strings.Replace(file, "/", "", 1)
	data, err := front.Asset(cFile)
	if err != nil {
		log.Println(err)
	}

	log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", r.UserAgent())
	/* отправить его клиенту */
	contentType := mime.TypeByExtension(path.Ext(cFile))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Server", version.Version)
	w.Write(data)
}

func (rt *Router) ApiHandler(w http.ResponseWriter, r *http.Request) {
	var request apiRequest
	var ok bool

	rt.nodes.RLock()
	defer rt.nodes.RUnlock()

	defer r.Body.Close()
	remoteIP := helpers.GetIP(r.RemoteAddr, r.Header.Get("X-Real-IP"), r.Header.Get("X-Forwarded-For"))

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Server", version.Version)

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

			nresp, err := rt.getNodes()

			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}

			j, _ := json.Marshal(nresp)
			w.Write(j)
		}

	case "get_indices":
		{
			response, err := rt.doGet(rt.conf.Elastic.Host + "_cat/indices/restored*?s=i&format=json")
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}

			log.Println(rt.nodes.nlist)
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

			status_response, err := rt.doGet(rt.conf.Elastic.Host + "_snapshot/" + repo + "/" + snap + "/_status")
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(remoteIP, "\t", r.Method, "\t", r.URL.Path, "\t", request.Action, "\t", 500, "\t", err.Error(), "\t", r.UserAgent())
				return
			}
			var status snapStatus // map[string]interface{}
			_ = json.Unmarshal(status_response, &status)
			log.Printf("%v", status)

			req := map[string]interface{}{
				"ignore_unavailable":   false,
				"include_global_state": false,
				"include_aliases":      false,
				"rename_pattern":       "(.+)",
				"rename_replacement":   "restored_$1",
				"indices":              request.Values["indices"],
				"index_settings":       map[string]interface{}{"index.number_of_replicas": 0},
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

func (rt *Router) getNodes() ([]singleNode, error) {

	var nresp []singleNode

	response, err := rt.doGet(rt.conf.Elastic.Host + "_cat/nodes?format=json&bytes=b&h=ip,name,dt,du,dup,d")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &rt.nodes.nlist)
	if err != nil {
		return nil, err
	}

	for i, n := range rt.nodes.nlist {
		rt.nodes.nlist[i].DiskFree = helpers.Atoi(n.D)
		nresp = append(nresp, singleNode{
			Ip:   rt.nodes.nlist[i].Ip,
			Name: rt.nodes.nlist[i].Name,
			Dt:   fmt.Sprintf("%dGb", helpers.Atoi(n.Dt)/(1024*1024*1024)),
			Dup:  rt.nodes.nlist[i].Dup,
		})
	}

	return nresp, nil

}
