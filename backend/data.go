package backend

import (
	"encoding/json"
	"bytes"
	"encoding/binary"
	"errors"

)

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

// Object Manipulation Section

func GetObject(bucket, key string, buckets *Databases) (*Meta, []byte, error){
	db := buckets.GetNoCreate(bucket)
	if db == nil {
		return nil, nil, nil
	}

	encodedData, err := db.Get(LReadOptions, []byte(key))
	if encodedData == nil {
		return nil, nil, nil
	}

	if err != nil {
		return nil, nil, err
	}

	meta, data, err := DecodeData(encodedData)
	if err != nil {
		return nil, nil, err
	}
	return meta, data, nil
}