package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"levelupdb/backend"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

type JSONIndexes struct {
	Keys []string `json:"keys"`
}

func secondaryIndex(w http.ResponseWriter, req *http.Request, bucket string, indexField string, startValue string, endValue string) {
	var r JSONIndexes
	// TODO: refactor this.
	// The index operation should be moved to backend
	if indexField == "$key" {
		keys, err := database.GetKeysRange(bucket, startValue, endValue)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		r.Keys = keys
		if d, err := json.Marshal(r); err == nil {
			w.Write(d)
		} else {
			w.WriteHeader(500)
		}
		return
	} else if indexField == "$bucket" {
		// TODO: do we care about start and end value? Riak seems to care if an
		// end value is thrown into this
		keys, err := database.GetAllKeys(bucket)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		r.Keys = keys
		if d, err := json.Marshal(r); err == nil {
			w.Write(d)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	indexDb := indexDatabase.GetBucketNoCreate(bucket)
	if indexDb == nil {
		w.WriteHeader(404)
		return
	}
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

// Algorithm is BFS. Stolen from Wikipedia :)
func walkLink(w http.ResponseWriter, req *http.Request, bucket string, key string, walks []string) {
	if len(walks) < 1 {
		w.WriteHeader(400)
		return
	}

	meta, _, err := database.GetObject(bucket, key)
	if err != nil {
		w.WriteHeader(400)
		return
	} else if meta == nil {
		w.WriteHeader(404)
		return
	}

	metaToExplore := list.New()
	metaToExplore.PushBack(meta)

	multipartBuffer := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(multipartBuffer)

	for i, walk := range walks {
		phase := strings.Split(walk, ",")
		if len(phase) != 3 {
			w.WriteHeader(400)
			return
		}

		phasepartBuffer := new(bytes.Buffer)
		phasepartWriter := multipart.NewWriter(phasepartBuffer)

		nextMeta := list.New()
		for node := metaToExplore.Front(); node != nil; node = node.Next() {
			meta := node.Value.(*backend.Meta)
			links := backend.QueryLinks(meta.Links, phase[0], phase[1])
			for _, link := range links {
				meta, body, err := database.GetObjectFromLink(link)
				if err != nil {
					w.WriteHeader(500)
					return
				} else if meta == nil {
					continue
				}

				if i+1 == len(walks) || phase[2] == "1" {
					partHeader := make(http.Header)
					meta.ToHeaders(partHeader, bucket)
					partHeader.Set("Location", fmt.Sprintf("/buckets/%s/keys/%s", link.Bucket, link.Key))
					wr, err := phasepartWriter.CreatePart(textproto.MIMEHeader(partHeader))
					if err != nil {
						w.WriteHeader(500)
						return
					}
					wr.Write(body)
				}

				if i+1 < len(walks) {
					nextMeta.PushBack(meta)
				}

			}
		}
		metaToExplore = nextMeta
		phaseHeader := make(textproto.MIMEHeader)
		phaseHeader.Set("Content-Type", "multipart/mixed; boundary="+phasepartWriter.Boundary())
		phasepartWriter.Close()
		wr, err := multipartWriter.CreatePart(phaseHeader)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		wr.Write(phasepartBuffer.Bytes())
	}

	w.Header().Set("Content-Type", "multipart/mixed; boundary="+multipartWriter.Boundary())
	multipartWriter.Close()
	w.Write(multipartBuffer.Bytes())
}

func mapred(w http.ResponseWriter, req *http.Request) {
}
