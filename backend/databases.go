package backend

import (
	"os"
	"io/ioutil"
	"path"
	"strings"
	"github.com/jmhodges/levigo"
	"bytes"
)

type Database struct {
	DBMap        map[string]*levigo.DB
	BaseLocation string
	IndexDatabase *Database
}

var LReadOptions *levigo.ReadOptions
var LWriteOptions *levigo.WriteOptions

func Initialize() {
	InitializeLeveldbOptions()
	InitializeLinkRegexp()
}

func InitializeLeveldbOptions() {
	LReadOptions = levigo.NewReadOptions()
	LWriteOptions = levigo.NewWriteOptions()
}

// Will panic if there is a problem with the database.
// Should only be called on server initialization.
func NewDatabase(databaseLocation string) *Database {
	buckets := new(Database)
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
		opts.SetCache(levigo.NewLRUCache(4194304))
		buckets.DBMap[file.Name()], err = levigo.Open(path.Join(databaseLocation, file.Name()), opts)
		if err != nil {
			panic(err)
		}
	}
	return buckets
}

func (buckets *Database) GetBucket(name string) (*levigo.DB, error) {
	if db, ok := buckets.DBMap[name]; ok {
		return db, nil
	}
	opts := levigo.NewOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCache(levigo.NewLRUCache(4194304))
	var err error
	buckets.DBMap[name], err = levigo.Open(path.Join(buckets.BaseLocation, name), opts)
	return buckets.DBMap[name], err
}

func (buckets *Database) GetBucketNoCreate(name string) *levigo.DB {
	if db, ok := buckets.DBMap[name]; ok {
		return db
	}
	return nil
}

func (buckets *Database) DestroyBucket(name string) error {
	if _, ok := buckets.DBMap[name]; ok {
		delete(buckets.DBMap, name)
		return levigo.DestroyDatabase(name, levigo.NewOptions())
	}
	return nil
}

func (buckets *Database) GetAllBucketNames() ([]string, error) {
	fileinfos, err := ioutil.ReadDir(buckets.BaseLocation)
	if err != nil {
		return nil, err
	}
	bucketNames := make([]string, len(fileinfos) - 1)
	for _, info := range fileinfos {
		name := info.Name()
		if name != "_indexes" && name != "" {
			bucketNames = append(bucketNames, name)
		}
	}
	return bucketNames, nil
}

func (buckets *Database) GetKeysRange(bucket, start, end string) ([]string, error) {
	if db, ok := buckets.DBMap[bucket]; ok {
		keys := make([]string, 0)
		it := db.NewIterator(LReadOptions) // TODO: Refactor with GetAllKeys
		it.Seek([]byte(start))
		var check func(*levigo.Iterator) bool
		if len(end) == 0 {
			check = func(*levigo.Iterator) bool {
				return true
			}
		} else {
			bend := []byte(end)
			check = func(*levigo.Iterator) bool {
				return bytes.Compare(bend, it.Key()) < 0
			}
		}
		for it = it; it.Valid() && check(it); it.Next() {
			keys = append(keys, string(it.Key()))
		}

		err := it.GetError()
		if err != nil {
			return nil, err
		} else {
			return keys, err
		}
	}
	return make([]string, 0), nil
}

func (buckets *Database) IsBucketEmpty(bucket string) bool {
	if db, ok := buckets.DBMap[bucket]; ok {
		it := db.NewIterator(LReadOptions)
		it.SeekToFirst()
		return !it.Valid()
	}
	return true
}

// TODO: implement auto bucket garbage collection
// TODO: test destroy database and is bucket empty

func (buckets *Database) GetAllKeys(bucket string) ([]string, error) {
	if db, ok := buckets.DBMap[bucket]; ok {
		keys := make([]string, 0)
		it := db.NewIterator(LReadOptions)
		it.SeekToFirst()
		for it = it; it.Valid(); it.Next() {
			keys = append(keys, string(it.Key()))
		}
		err := it.GetError()
		if err != nil {
			return nil, err
		} else {
			return keys, nil
		}
	}
	return make([]string, 0), nil
}

func (buckets *Database) StreamAllKeys(bucket string, keys chan<- string) {
	if db, ok := buckets.DBMap[bucket]; ok {
		it := db.NewIterator(LReadOptions)
		it.SeekToFirst()
		for it = it; it.Valid(); it.Next() {
			keys <- string(it.Key())
		}

		if err := it.GetError(); err == nil {
			keys <- ""
		} // TODO: What about error?
	}
}