package main

import (
	"net/http"
	"log"
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

const VERSION = "0.1"

var MainLogger *log.Logger

func standardHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		MainLogger.Println("-", request.RemoteAddr, "-", request.Method, request.URL)
		header := w.Header()
		header.Add("Server", "levelupdb/"+VERSION)
		fn(w, request)
	}
}


type Config struct {
	DatabaseLocation string
	Logging string
}

func main() {

	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(fmt.Sprintln("Config file error: ", err))
	}

	config := new(Config)
	err = json.Unmarshal(data, config)

	if err != nil {
		panic(fmt.Sprintln("Config file error: ", err))
	}

	var writer io.Writer
	if config.Logging == "stdout" {
		writer = os.Stdout
	} else if config.Logging == "none" {
		writer = ioutil.Discard
	} else {
		writer = os.Stdout // TODO: change this to files.
	}

	MainLogger = log.New(writer, "[levelupdb " + VERSION + "] ", log.Ldate | log.Ltime)

	// Server Operations
	http.HandleFunc("/ping", standardHandler(ping))
	http.HandleFunc("/", standardHandler(listResources))

	MainLogger.Println("NOTICE: Server started")
	http.ListenAndServe(":8198", nil)
}
