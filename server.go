package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmhodges/levigo"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
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
var Buckets *Databases
var IndexDbs *Databases

var LReadOptions *levigo.ReadOptions
var LWriteOptions *levigo.WriteOptions

func main() {

	initializeLinkRegexp()

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
	Buckets = GetAllBuckets(DBConfig.DatabaseLocation)
	IndexDbs = GetAllBuckets(path.Join(DBConfig.DatabaseLocation, "_indexes"))
	LReadOptions = levigo.NewReadOptions()
	LWriteOptions = levigo.NewWriteOptions()

	// Server Operations
	http.HandleFunc("/ping", standardHandler(ping))
	http.HandleFunc("/", standardHandler(listResources))
	http.HandleFunc("/buckets/", standardHandler(bucketsOps))
	http.HandleFunc("/stats", standardHandler(stats))

	MainLogger.Println("NOTICE: Server started. Serving port " + DBConfig.HttpPort)
	http.ListenAndServe(":"+DBConfig.HttpPort, nil)
}
