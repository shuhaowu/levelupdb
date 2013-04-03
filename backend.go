package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/jmhodges/levigo"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

type Databases struct {
	DBMap        map[string]*levigo.DB
	BaseLocation string
}

// Will panic if there is a problem with the database.
// Should only be called on server initialization.
func GetAllBuckets(databaseLocation string) *Databases {
	buckets := new(Databases)
	buckets.DBMap = make(map[string]*levigo.DB)
	buckets.BaseLocation = databaseLocation

	os.MkdirAll(databaseLocation, 0755)

	files, err := ioutil.ReadDir(databaseLocation)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() || strings.HasPrefix(file.Name(), "_") {
			continue
		}
		opts := levigo.NewOptions()
		opts.SetCreateIfMissing(true)
		opts.SetCache(levigo.NewLRUCache(1 << 30))
		buckets.DBMap[file.Name()], err = levigo.Open(path.Join(databaseLocation, file.Name()), opts)
		if err != nil {
			panic(err)
		}
	}
	return buckets
}

func (buckets *Databases) Get(name string) (*levigo.DB, error) {
	if db, ok := buckets.DBMap[name]; ok {
		return db, nil
	}
	opts := levigo.NewOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCache(levigo.NewLRUCache(1 << 30))
	var err error
	buckets.DBMap[name], err = levigo.Open(path.Join(buckets.BaseLocation, name), opts)
	return buckets.DBMap[name], err
}

func (buckets *Databases) GetNoCreate(name string) *levigo.DB {
	if db, ok := buckets.DBMap[name]; ok {
		return db
	}
	return nil
}

func (buckets *Databases) Purge(name string) error {
	if db, ok := buckets.DBMap[name]; ok {
		db.Close() // TODO: use DestroyDatabase instead
		delete(buckets.DBMap, name)
		return os.RemoveAll(path.Join(buckets.BaseLocation, name))
	}
	return nil
}

type Meta struct {
	Indexes     [][2]string       `json:"I"`
	Links       string            `json:"L"`
	Meta        map[string]string `json:"M"`
	ContentType string            `json:"C"`
}

func MetaFromRequest(req *http.Request) (*Meta, error) {
	meta := new(Meta)

	meta.Links = req.Header.Get("Link")
	meta.ContentType = req.Header.Get("Content-Type")
	for headerKey, headerValue := range req.Header {
		headerValueLength := len(headerValue)
		if strings.HasPrefix(headerKey, "X-Riak-Index-") && headerValueLength > 0 {
			indexKey := strings.ToLower(headerKey[13:]) // case insenstive because go convert the first character into caps?
			index := [2]string{indexKey, headerValue[0]}
			meta.Indexes = append(meta.Indexes, index)
		}

		if strings.HasPrefix(headerKey, "X-Riak-Meta-") && headerValueLength > 0 {
			metaKey := strings.ToLower(headerKey[12:]) // same reason as above.
			meta.Meta[metaKey] = headerValue[0]
		}
	}

	return meta, nil
}

func MetaToHeaders(headers http.Header, meta *Meta) {
	if meta.Links != "" {
		headers.Add("Link", meta.Links)
	}
	headers.Add("Content-Type", meta.ContentType)
	for _, index := range meta.Indexes {
		headers.Add("X-Riak-Index-"+index[0], index[1])
	}

	for k, v := range meta.Meta {
		headers.Add("X-Riak-Meta-"+k, v)
	}
	headers.Add("X-Riak-Vclock", "Yay02966e9d038d6332eea23012217f8c4b521eaf92==")
}

func appendDataKey(keys []byte, key []byte) []byte {
	if len(keys) > 0 {
		keys = append(keys, byte(9))
	}

	return append(keys, key...)
}

func removeDataKey(keys []byte, key []byte) []byte {
	kStart := bytes.Index(keys, key)
	kEnd := kStart + len(key)

	if kEnd < len(keys)-1 {
		kEnd++
	} else {
		if kStart > 0 {
			kStart--
		}
	}

	return append(append([]byte{}, keys[:kStart]...), keys[kEnd:]...)
}

func decodeDataKeys(keys []byte) []string {
	return strings.Split(string(keys), string(byte(9)))
}

func AddIndex(index [2]string, key []byte, indexDb *levigo.DB, wb *levigo.WriteBatch) error {
	searchKey := []byte(index[0] + "~" + index[1])
	keys, err := indexDb.Get(LReadOptions, searchKey)
	if err != nil {
		return err
	}
	keys = appendDataKey(keys, key)
	wb.Put(searchKey, keys)
	return nil
}

// TODO: refactor with above.
func RemoveIndex(index [2]string, key []byte, indexDb *levigo.DB, wb *levigo.WriteBatch) error {
	searchKey := []byte(index[0] + "~" + index[1])
	keys, err := indexDb.Get(LReadOptions, searchKey)
	if err != nil {
		return err
	}
	keys = removeDataKey(keys, key)
	if len(keys) > 0 {
		wb.Put(searchKey, keys)
	} else {
		wb.Delete(searchKey)
	}
	return nil
}

// This method takes raw indexes (unsplit and whatnot) and computes the
// diff indexes that's split.
// So note that it in takes a slice of an array of 2 which consists of the
// field_key and all values and returns a slice of arrays of two of field and value.
// NOTE: THIS IS REALLY INEFFICIENT.
func ComputeIndexesDiff(newIndexes [][2]string, oldIndexes [][2]string) (added [][2]string, deleted [][2]string) {
	newIndexesMap := make(map[[2]string]bool)
	oldIndexesMap := make(map[[2]string]bool)

	for _, allIndexes := range newIndexes {
		splitted := strings.Split(allIndexes[1], ",") // BUG?: What if a value has a comma in it?
		for _, ivalue := range splitted {
			newIndexesMap[[2]string{allIndexes[0], ivalue}] = true
		}
	}

	for _, allIndexes := range oldIndexes {
		splitted := strings.Split(allIndexes[1], ",") // BUG?: What if a value has a comma in it?
		for _, ivalue := range splitted {
			index := [2]string{allIndexes[0], ivalue}
			_, exists := newIndexesMap[index]
			if !exists {
				deleted = append(deleted, index)
				continue // We don't need to store this.
			}
			oldIndexesMap[index] = true
		}
	}

	for index, _ := range newIndexesMap {
		_, exists := oldIndexesMap[index]
		if !exists {
			added = append(added, index)
		}
	}

	return
}

func GenerateWriteBatchForIndexes(added, deleted [][2]string, key string, indexDb *levigo.DB) (*levigo.WriteBatch, error) {
	wb := levigo.NewWriteBatch()
	bkey := []byte(key)
	for _, index := range added {
		if err := AddIndex(index, bkey, indexDb, wb); err != nil {
			return nil, err
		}
	}

	for _, index := range deleted {
		if err := RemoveIndex(index, bkey, indexDb, wb); err != nil {
			return nil, err
		}
	}
	return wb, nil
}

func EncodeData(meta *Meta, data []byte) ([]byte, error) {
	metaString, err := json.Marshal(meta)

	if err != nil {
		return nil, err
	}

	length := int32(len(metaString))
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, length)

	if err != nil {
		return nil, err
	}

	buf.Write(metaString)
	buf.Write(data)
	return buf.Bytes(), nil
}

func DecodeData(data []byte) (*Meta, []byte, error) {
	if len(data) < 5 {
		return nil, nil, errors.New("Data length must be greater than 4!")
	}

	var length int32
	lbuf := bytes.NewBuffer(data[:4])
	err := binary.Read(lbuf, binary.BigEndian, &length)
	if err != nil {
		return nil, nil, err
	}

	meta := new(Meta)
	length += 4 // Should never overflow, right?
	err = json.Unmarshal(data[4:length], meta)

	if err != nil {
		return nil, nil, err
	}

	// Returning a slice from the original data array should be fine, right?
	return meta, data[length:], nil
}
