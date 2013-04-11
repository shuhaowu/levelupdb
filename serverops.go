package main

import (
	"encoding/json"
	"net/http"
)

func ping(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}

type Resources struct {
	Riak_kv_wm_buckets     string `json:"riak_kv_wm_buckets"`
	Riak_kv_wm_index       string `json:"riak_kv_wm_index"`
	Riak_kv_wm_keylist     string `json:"riak_kv_wm_keylist"`
	Riak_kv_wm_link_walker string `json:"riak_kv_wm_link_walker"`
	Riak_kv_wm_mapred      string `json:"riak_kv_wm_mapred"`
	Riak_kv_wm_object      string `json:"riak_kv_wm_object"`
	Riak_kv_wm_ping        string `json:"riak_kv_wm_ping"`
	Riak_kv_wm_props       string `json:"riak_kv_wm_props"`
	Riak_kv_wm_stats       string `json:"riak_kv_wm_stats"`
}

var resources Resources = Resources{
	Riak_kv_wm_buckets:     "/buckets",
	Riak_kv_wm_index:       "/buckets",
	Riak_kv_wm_keylist:     "/buckets",
	Riak_kv_wm_link_walker: "/buckets",
	Riak_kv_wm_mapred:      "/mapred",
	Riak_kv_wm_object:      "/buckets",
	Riak_kv_wm_ping:        "/ping",
	Riak_kv_wm_props:       "/buckets",
	Riak_kv_wm_stats:       "/stats",
}

func listResources(w http.ResponseWriter, req *http.Request) {
	header := req.Header
	if header.Get("Accept") == "application/json" {
		data, err := json.Marshal(resources)
		if err != nil {
			mainLogger.Fatalln("Decoding resources json failed:", resources)
		} else {
			w.Write(data)
		}
	} else {
		w.WriteHeader(404)
		if len(req.URL.Path) == 1 {
			w.Write([]byte("We're not participating in the html nonsense.\n"))
		}
	}
}

func stats(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("{\"riak_kv_version\":\"1.3.1\",\"riak_api_version\":\"1.3.1\"}"))
}
