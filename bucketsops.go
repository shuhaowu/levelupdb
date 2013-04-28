package main

import (
	"net/http"
	"strings"
	"encoding/json"
)

const lenPath = len("/buckets/")

// This function just delegates.
func bucketsOps(w http.ResponseWriter, req *http.Request) {
	remainingUrl := req.URL.Path[lenPath:]
	if len(remainingUrl) == 0 {
		listBuckets(w, req)
	} else {
		splitted := strings.Split(remainingUrl, "/")
		length := len(splitted)
		if length == 2 && splitted[1] == "keys" {
			switch {
			case req.Method == "POST":
				storeObject(w, req, splitted[0], "")
			case req.Method == "GET":
				keys := req.URL.Query().Get("keys")
				if keys == "true" {
					listKeys(w, req, splitted[0])
				} else if keys == "stream" {
					// TODO: Stream keys
				}
			}
		} else if length == 3 && splitted[1] == "keys" {
			bucket := splitted[0]
			key := splitted[2]
			switch {
			case req.Method == "GET":
				fetchObject(w, req, bucket, key)
			case req.Method == "PUT" || req.Method == "POST":
				storeObject(w, req, bucket, key)
			case req.Method == "DELETE":
				deleteObject(w, req, bucket, key)
			}
		} else if length >= 4 && splitted[1] == "index" {
			bucket := splitted[0]
			indexField := splitted[2]
			startValue := splitted[3]
			endValue := ""
			if length == 5 {
				endValue = splitted[4]
			}
			secondaryIndex(w, req, bucket, indexField, startValue, endValue)
		} else if length >= 4 && splitted[1] == "keys" { // has to be links, right?
			walkLink(w, req, splitted[0], splitted[2], splitted[3:])
		} else {
			w.WriteHeader(404)
		}
	}
}

type allBuckets struct {
	Buckets []string `json:"buckets"`
}

func listBuckets(w http.ResponseWriter, req *http.Request) {
	var all allBuckets
	buckets, err := database.GetAllBucketNames()
	all.Buckets = buckets
	if err != nil {
		mainLogger.Println("ERROR: Getting all databases name failed with", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(all)
	if err != nil {
		w.WriteHeader(500)
		return
	} else {
		w.Write(data)
	}
}

type allKeys struct {
	Keys []string `json:"keys"`
}

func listKeys(w http.ResponseWriter, req *http.Request, bucket string) {
	var all allKeys
	keys, err := database.GetAllKeys(bucket)
	all.Keys = keys
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if data, err := json.Marshal(all) ; err == nil{
		w.Write(data)
	} else {
		w.WriteHeader(500)
	}
}
