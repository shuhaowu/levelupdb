package main

import (
	//	"encoding/json"
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
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
	db := Buckets.GetNoCreate(bucket)
	if db == nil {
		w.WriteHeader(404)
		return
	}

	encodedData, err := db.Get(LReadOptions, []byte(key))
	if encodedData == nil {
		w.WriteHeader(404)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Getting database error", err)
		return
	}

	meta, data, err := DecodeData(encodedData)
	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Decoding data error", err)
		return
	}

	MetaToHeaders(w.Header(), meta)
	w.Write(data)
}

func storeObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		MainLogger.Printf("Error: Error reading request body '%s'.", err)
		w.WriteHeader(400)
		return
	}
	meta, err := MetaFromRequest(req)

	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Meta construction failed on header", req.Header)
		return
	}

	encodedData, err := EncodeData(meta, data)
	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Data encoding failed on data", string(data))
		return
	}

	db, err := Buckets.Get(bucket)
	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Getting bucket failed with bucket", bucket)
		return
	}

	created := false
	if key == "" {
		key, err = GenUUID()
		if err != nil {
			w.WriteHeader(500)
			MainLogger.Println("ERROR: Generating UUID Failed.")
			return
		}
		created = true
	}

	err = db.Put(LWriteOptions, []byte(key), encodedData)

	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Writing data failed with key", key, "and data", encodedData)
		return
	}

	returnbody := req.URL.Query().Get("returnbody") == "true"

	if created {
		w.Header().Add("Location", "/buckets/"+bucket+"/keys/"+key)
	}

	if returnbody {
		MetaToHeaders(w.Header(), meta)
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
	db := Buckets.GetNoCreate(bucket)
	if db == nil {
		w.WriteHeader(404)
		return
	}

	err := db.Delete(LWriteOptions, []byte(key))
	if err != nil {
		w.WriteHeader(500)
		MainLogger.Println("ERROR: Deletion failed with key", key)
		return
	}
	w.WriteHeader(204)
}
