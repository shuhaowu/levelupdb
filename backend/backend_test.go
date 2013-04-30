package backend

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestEncodingDecoding(t *testing.T) {
	meta := new(Meta)
	meta.Indexes = make([][2]string, 1)
	meta.Meta = make(map[string]string)

	meta.Meta["test"] = "yay"
	meta.Indexes[0] = [2]string{"field", "value"}

	metaString, _ := json.Marshal(meta)
	metaLength := len(metaString)

	myData := []byte{1, 2, 3}

	data, err := EncodeData(meta, myData)
	if err != nil {
		t.Fatal(err)
	}

	var length int32
	lbuf := bytes.NewBuffer(data[:4])
	err = binary.Read(lbuf, binary.BigEndian, &length)

	if err != nil {
		t.Fatal(err)
	}

	if length != int32(metaLength) {
		t.Fatal("Encode: Length does not equal", length, "!=", metaLength)
	}

	if !bytes.Equal(data[4:4+metaLength], metaString) {
		t.Fatal("Encode: Meta does not equal!")
	}

	if !bytes.Equal(data[4+metaLength:], myData) {
		t.Fatal("Encode: Data does not equal!")
	}

	decodedMeta, decodedData, err := DecodeData(data)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decodedData, myData) {
		t.Fatal("Decode: Data does not equal!")
	}

	for k, v := range meta.Meta {
		if decodedMeta.Meta[k] != v {
			t.Fatal("Decode: Meta does not equal!")
		}
	}

	for i, value := range meta.Indexes {
		if value[0] != decodedMeta.Indexes[i][0] || value[1] != decodedMeta.Indexes[i][1] {
			t.Fatal("Decode: Meta does not equal!")
		}
	}
}

func TestKeyBucket(t *testing.T) {
	keys := [][]byte{[]byte("yay"), []byte("woo!"), []byte("meow")}
	expectedResult := bytes.Join(keys, []byte{9})

	var actualResult []byte
	actualResult = appendDataKey(actualResult, keys[0])
	actualResult = appendDataKey(actualResult, keys[1])
	actualResult = appendDataKey(actualResult, keys[2])

	if !bytes.Equal(expectedResult, actualResult) {
		t.Fatal("KeyBucket: Append failed")
	}

	removedResult := removeDataKey(actualResult, keys[0])
	expectedResult = bytes.Join([][]byte{[]byte("woo!"), []byte("meow")}, []byte{9})
	if !bytes.Equal(expectedResult, removedResult) {
		t.Fatal("KeyBucket: Remove failed (start)")
	}

	removedResult = removeDataKey(actualResult, keys[1])
	expectedResult = bytes.Join([][]byte{[]byte("yay"), []byte("meow")}, []byte{9})
	if !bytes.Equal(expectedResult, removedResult) {
		t.Fatal("KeyBucket: Remove failed (middle)")
	}

	removedResult = removeDataKey(actualResult, keys[2])
	expectedResult = bytes.Join([][]byte{[]byte("yay"), []byte("woo!")}, []byte{9})
	if !bytes.Equal(expectedResult, removedResult) {
		t.Fatal("KeyBucket: Remove failed (end)")
	}

	decodedKeys := DecodeDataKeys(actualResult)
	for i, k := range decodedKeys {
		if !bytes.Equal(keys[i], []byte(k)) {
			t.Fatal("KeyBucket: Decode failed")
		}
	}
}

func TestLinkDecode(t *testing.T) {
	InitializeLinkRegexp()
	links := `</riak/list/1>; riaktag="previous"`
	link := ParseLink(links)
	if link.bucket != "list" {
		t.Fatal("Link: bucket decode failure")
	}

	if link.key != "1" {
		t.Fatal("Link: key decode failure")
	}

	if link.tag != "previous" {
		t.Fatal("Link: tag decode failure")
	}
}

func TestQueryLinks(t *testing.T) {
	InitializeLinkRegexp()
	links := `</riak/list/1>; riaktag="previous", </riak/list/3>; riaktag="next", </riak/list2/1>; riaktag="next"`
	results := QueryLinks(links, "list", "previous")
	if len(results) != 1 {
		t.Fatal("QueryLinks: Results not length 1")
	}

	if results[0].bucket != "list" || results[0].tag != "previous" || results[0].key != "1" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	results = QueryLinks(links, "list", "_")
	if len(results) != 2 {
		t.Fatal("QueryLinks: Results not length 2")
	}

	if results[0].bucket != "list" || results[0].tag != "previous" || results[0].key != "1" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	if results[1].bucket != "list" || results[1].tag != "next" || results[1].key != "3" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	results = QueryLinks(links, "_", "next")
	if len(results) != 2 {
		t.Fatal("QueryLinks: Results not length 2")
	}

	if results[0].bucket != "list" || results[0].tag != "next" || results[0].key != "3" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	if results[1].bucket != "list2" || results[1].tag != "next" || results[1].key != "1" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	results = QueryLinks(links, "_", "_")

	if len(results) != 3 {
		t.Fatal("QueryLinks: Results not length 3")
	}

	if results[0].bucket != "list" || results[0].tag != "previous" || results[0].key != "1" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	if results[1].bucket != "list" || results[1].tag != "next" || results[1].key != "3" {
		t.Fatal("QueryLinks: Wrong result.")
	}

	if results[2].bucket != "list2" || results[2].tag != "next" || results[2].key != "1" {
		t.Fatal("QueryLinks: Wrong result.")
	}
}

