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
			indexKey := headerKey[13:] // we store this to save some more space...
			index := [2]string{indexKey, headerValue[0]}
			meta.Indexes = append(meta.Indexes, index)
		}

		if strings.HasPrefix(headerKey, "X-Riak-Meta-") && headerValueLength > 0 {
			metaKey := headerKey[12:]
			meta.Meta[metaKey] = headerValue[0]
		}
	}

	return meta, nil
}

func MetaToHeaders(headers http.Header, meta *Meta) {
	headers.Add("Link", meta.Links)
	headers.Add("Content-Type", meta.ContentType)
	for _, index := range meta.Indexes {
		headers.Add("X-Riak-Index-"+index[0], index[1])
	}

	for k, v := range meta.Meta {
		headers.Add("X-Riak-Meta-"+k, v)
	}
	headers.Add("X-Riak-Vclock", "Yay02966e9d038d6332eea23012217f8c4b521eaf92==")
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
