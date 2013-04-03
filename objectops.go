package main

import (
	//	"encoding/json"
	"io/ioutil"
	"net/http"
)

func fetchObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
}

func storeObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
	_, err := ioutil.ReadAll(req.Body)
	if err != nil {
		MainLogger.Printf("WARNING: Error reading request body '%s'. Ignored.", err)
	}

}

func deleteObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
}
