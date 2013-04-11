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
		mainLogger.Println("-", request.RemoteAddr, "-", request.Method, request.URL.Path)
	}
}

type Config struct {
	DatabaseLocation string
	Logging          string
	HttpPort         string
}

func initializeConfig() *Config {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(fmt.Sprintln("Config file error: ", err))
	}

	config := new(Config)
	err = json.Unmarshal(data, config)

	if err != nil {
		panic(fmt.Sprintln("Config file error: ", err))
	}
	return config
}

func initializeLogger() *log.Logger {
	var writer io.Writer
	if globalConfig.Logging == "stdout" {
		writer = os.Stdout
	} else if globalConfig.Logging == "none" {
		writer = ioutil.Discard
	} else {
		writer = os.Stdout // TODO: change this to files.
	}

	return log.New(writer, "[levelupdb "+VERSION+"] ", log.Ldate|log.Ltime)
}

var mainLogger *log.Logger
var globalConfig *Config
var database *backend.Database
var indexDatabase *backend.Database

func main() {
	backend.Initialize()
	globalConfig = initializeConfig()
	mainLogger = initializeLogger()

	database = backend.NewDatabase(globalConfig.DatabaseLocation)
	indexDatabase = backend.NewDatabase(path.Join(globalConfig.DatabaseLocation, "_indexes"))

	// Server Operations
	http.HandleFunc("/ping", standardHandler(ping))
	http.HandleFunc("/", standardHandler(listResources))
	http.HandleFunc("/buckets/", standardHandler(bucketsOps))
	http.HandleFunc("/stats", standardHandler(stats))

	mainLogger.Println("NOTICE: Server started. Serving port " + globalConfig.HttpPort)
	http.ListenAndServe(":"+globalConfig.HttpPort, nil)
}
