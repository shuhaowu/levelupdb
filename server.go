/*
 * This file is part of levelupdb.
 *
 * levelupdb is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * levelupdb is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with levelupdb.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"levelupdb/backend"
	"log"
	"net/http"
	"os"
	"path"
)

const VERSION = "0.1"
const SERVER_STRING = "levelupdb/" + VERSION + " (someone painted it purple)"

func standardHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		header := w.Header()
		header.Add("Server", SERVER_STRING)
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
	database.IndexDatabase = indexDatabase

	// Server Operations
	http.HandleFunc("/ping", standardHandler(ping))
	http.HandleFunc("/", standardHandler(listResources))
	http.HandleFunc("/buckets/", standardHandler(bucketsOps))
	http.HandleFunc("/buckets", standardHandler(listBuckets))
	http.HandleFunc("/stats", standardHandler(stats))

	mainLogger.Println("NOTICE: Server started. Serving port " + globalConfig.HttpPort)
	http.ListenAndServe(":"+globalConfig.HttpPort, nil)
}
