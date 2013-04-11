package main

import (
	"bytes"
	"encoding/json"
	"levelupdb/backend"
	"net/http"
	"strings"
)

type JSONIndexes struct {
	Keys []string `json:"keys"`
}

func secondaryIndex(w http.ResponseWriter, req *http.Request, bucket string, indexField string, startValue string, endValue string) {
	indexDb := indexDatabase.GetBucketNoCreate(bucket)
	if indexDb == nil {
		w.WriteHeader(404)
		return
	}
	var r JSONIndexes
	searchKey := []byte(indexField + "~" + startValue)
	if endValue == "" {
		keys, err := indexDb.Get(backend.LReadOptions, searchKey)
		if err != nil {
			w.WriteHeader(500)
			mainLogger.Println("ERROR: Getting index values failed with", searchKey)
			return
		}

		r.Keys = backend.DecodeDataKeys(keys)
	} else {
		endSearchKey := []byte(indexField + "~" + endValue)
		it := indexDb.NewIterator(backend.LReadOptions)
		it.Seek(searchKey)
		for it = it; it.Valid(); it.Next() {
			if bytes.Compare(it.Key(), endSearchKey) > 0 {
				break
			}
			r.Keys = append(r.Keys, backend.DecodeDataKeys(it.Value())...)
		}
	}
	d, err := json.Marshal(r)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: JSON decode failed with ", r.Keys)
		return
	}
	w.Write(d)
}

func walkLink(w http.ResponseWriter, req *http.Request, bucket string, key string, walks []string) {
	for _, phasestr := range walks {
		phase := strings.Split(phasestr, ",")
		if len(phase) != 3 {
			w.WriteHeader(400)
			return
		}
	}
}

func mapred(w http.ResponseWriter, req *http.Request) {

}
