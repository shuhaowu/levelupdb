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
package backend

import (
	"encoding/json"
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
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

func (database *Database) GetObject(bucket, key string) (*Meta, []byte, error) {
	db := database.GetBucketNoCreate(bucket)
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

func (database *Database) StoreObject(bucket, key string, meta *Meta, data []byte) error {
	db, err := database.GetBucket(bucket)
	if err != nil {
		return err
	}

	indexDb, err := database.IndexDatabase.GetBucket(bucket)
	if err != nil {
		return err
	}

	encodedData, err := EncodeData(meta, data)
	if err != nil {
		return err
	}

	bkey := []byte(key)
	oldData, err := db.Get(LReadOptions, bkey)
	if err != nil {
		return err
	}

	var oldIndexes [][2]string
	if oldData != nil {
		oldMeta, _, err := DecodeData(oldData)
		if err != nil {
			return err
		}
		oldIndexes = oldMeta.Indexes
	}

	if err = db.Put(LWriteOptions, bkey, encodedData); err != nil {
		return err
	}

	addedIndexes, deletedIndexes := ComputeIndexesDiff(meta.Indexes, oldIndexes)
	wb, err := GenerateWriteBatchForIndexes(addedIndexes, deletedIndexes, key, indexDb)
	if err != nil {
		return err
	}

	if err := indexDb.Write(LWriteOptions, wb); err != nil {
		return err
	}

	return nil
}

func (database *Database) DeleteObject(bucket, key string) (int, error){
	db := database.GetBucketNoCreate(bucket)
	if db == nil {
		return 404, nil
	}

	bkey := []byte(key)
	encodedData, _ := db.Get(LReadOptions, bkey)
	if encodedData == nil {
		return 404, nil
	}

	err := db.Delete(LWriteOptions, bkey)
	if err != nil {
		return 500, err
	}

	meta, _, _ := DecodeData(encodedData)
	if meta != nil && len(meta.Indexes) > 0 {
		indexDb := database.IndexDatabase.GetBucketNoCreate(bucket)
		if indexDb != nil {
			var added [][2]string
			var removed [][2]string
			for _, indexes := range meta.Indexes {
				splitted := strings.Split(indexes[1], ",")
				for _, value := range splitted {
					removed = append(removed, [2]string{indexes[0], value})
				}
			}
			wb, err := GenerateWriteBatchForIndexes(added, removed, key, indexDb)
			if err != nil {
				return 500, err
			}

			if err = indexDb.Write(LWriteOptions, wb); err != nil {
				return 500, err
			}
		}
	}
	return 204, nil
}