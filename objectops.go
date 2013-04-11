package main

import (
	//	"encoding/json"
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strings"
	"levelupdb/backend"
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

	backend.MetaToHeaders(w.Header(), meta)
	w.Write(data)
}

func storeObject(w http.ResponseWriter, req *http.Request, bucket string, key string) {
	bkey := []byte(key)

	db, err := database.GetBucket(bucket)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Getting bucket failed with bucket", bucket)
		return
	}

	indexDb, err := indexDatabase.GetBucket(bucket)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Getting indexbucket failed with bucket", bucket)
		return
	}

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

	encodedData, err := backend.EncodeData(meta, data)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Data encoding failed on data", string(data))
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

	oldData, err := db.Get(backend.LReadOptions, bkey)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Getting previous data failed.")
		return
	}

	var oldIndexes [][2]string
	if oldData != nil {
		oldMeta, _, err := backend.DecodeData(oldData)
		if err != nil {
			w.WriteHeader(500)
			mainLogger.Println("ERROR: Decoding previous data error", err)
			return
		}
		oldIndexes = oldMeta.Indexes
	}

	addedIndexes, deletedIndexes := backend.ComputeIndexesDiff(meta.Indexes, oldIndexes)
	if err = db.Put(backend.LWriteOptions, []byte(key), encodedData); err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Writing data failed with key", key, "and data", encodedData)
		return
	}

	wb, err := backend.GenerateWriteBatchForIndexes(addedIndexes, deletedIndexes, key, indexDb)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Index writebatch generation failed.")
		return
	}

	if err := indexDb.Write(backend.LWriteOptions, wb); err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Index writes failed.")
		return
	}

	returnbody := req.URL.Query().Get("returnbody") == "true"

	if created {
		w.Header().Add("Location", "/buckets/"+bucket+"/keys/"+key)
	}

	if returnbody {
		backend.MetaToHeaders(w.Header(), meta)
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
	db := database.GetBucketNoCreate(bucket)
	if db == nil {
		w.WriteHeader(404)
		return
	}

	bkey := []byte(key)
	encodedData, _ := db.Get(backend.LReadOptions, bkey) // don't really care about error?
	if encodedData == nil {
		w.WriteHeader(404)
		return
	}

	err := db.Delete(backend.LWriteOptions, bkey)
	if err != nil {
		w.WriteHeader(500)
		mainLogger.Println("ERROR: Deletion failed with key", key)
		return
	}

	meta, _, _ := backend.DecodeData(encodedData)

	if meta != nil && len(meta.Indexes) > 0 {
		indexDb := indexDatabase.GetBucketNoCreate(bucket)
		if indexDb != nil {
			var added [][2]string
			var removed [][2]string
			for _, indexes := range meta.Indexes {
				splitted := strings.Split(indexes[1], ",")
				for _, value := range splitted {
					removed = append(removed, [2]string{indexes[0], value})
				}
			}
			wb, err := backend.GenerateWriteBatchForIndexes(added, removed, key, indexDb)
			if err != nil {
				w.WriteHeader(500)
				mainLogger.Println("ERROR: Index WriteBatch generation failed with key", key)
				return
			}

			if err = indexDb.Write(backend.LWriteOptions, wb); err != nil {
				w.WriteHeader(500)
				mainLogger.Println("ERROR: Index deletion failed with key", key)
				return
			}
		}
	}

	w.WriteHeader(204)
}
