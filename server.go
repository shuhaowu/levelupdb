package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"levelupdb/backend"
)

const VERSION = "0.1"

func standardHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		header := w.Header()
		header.Add("Server", "levelupdb/"+VERSION+" (someone painted it purple)")
		fn(w, request)
		MainLogger.Println("-", request.RemoteAddr, "-", request.Method, request.URL.Path)
	}
}

type Config struct {
	DatabaseLocation string
	Logging          string
	HttpPort         string
}

var MainLogger *log.Logger
var DBConfig *Config
var Buckets *backend.Databases
var IndexDbs *backend.Databases

func main() {
	backend.Initialize()

	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(fmt.Sprintln("Config file error: ", err))
	}

	DBConfig = new(Config)
	err = json.Unmarshal(data, DBConfig)

	if err != nil {
		panic(fmt.Sprintln("Config file error: ", err))
	}

	var writer io.Writer
	if DBConfig.Logging == "stdout" {
		writer = os.Stdout
	} else if DBConfig.Logging == "none" {
		writer = ioutil.Discard
	} else {
		writer = os.Stdout // TODO: change this to files.
	}

	MainLogger = log.New(writer, "[levelupdb "+VERSION+"] ", log.Ldate|log.Ltime)
	Buckets = backend.GetAllBuckets(DBConfig.DatabaseLocation)
	IndexDbs = backend.GetAllBuckets(path.Join(DBConfig.DatabaseLocation, "_indexes"))

	// Server Operations
	http.HandleFunc("/ping", standardHandler(ping))
	http.HandleFunc("/", standardHandler(listResources))
	http.HandleFunc("/buckets/", standardHandler(bucketsOps))
	http.HandleFunc("/stats", standardHandler(stats))

	MainLogger.Println("NOTICE: Server started. Serving port " + DBConfig.HttpPort)
	http.ListenAndServe(":"+DBConfig.HttpPort, nil)
}
