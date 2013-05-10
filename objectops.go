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
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"levelupdb/backend"
	"net/http"
)

// UUID from http://www.ashishbanerjee.com/home/go/go-generate-uuid
func GenUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil
}

func fetchObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
	meta, data, err := database.GetObject(bucket, key)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Getting object failed with err", err)
		return
	}

	if meta == nil {
		w.WriteHeader(404)
		return
	}

	meta.ToHeaders(w.Header(), bucket)
	w.Write(data)
}

func storeObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		mainLogger.Printf("Error: Error reading request body '%s'.", err)
		w.WriteHeader(400)
		return
	}

	meta, err := backend.MetaFromRequest(req)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Meta construction failed on header", req.Header)
		return
	}

	created := false
	if key == "" {
		if key, err = GenUUID(); err != nil {
			w.WriteHeader(500)
			mainLogger.Println("ERROR: Generating UUID Failed.")
			return
		}
		created = true
	}

	if err := database.StoreObject(bucket, key, meta, data); err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Backend store object failed with", err)
		return
	}

	returnbody := req.URL.Query().Get("returnbody") == "true"

	if created {
		w.Header().Add("Location", "/buckets/"+bucket+"/keys/"+key)
	}

	// This is ugly.
	if returnbody {
		meta.ToHeaders(w.Header(), bucket)
		if created {
			w.WriteHeader(201)
		}
		w.Write(data)
	} else {
		if created {
			w.WriteHeader(201)
		} else {
			w.WriteHeader(204)
		}
	}
}

func deleteObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
	code, err := database.DeleteObject(bucket, key)

	if err != nil {
		mainLogger.Println("ERROR: During delete...:w", err)
	}

	w.WriteHeader(code)
}
