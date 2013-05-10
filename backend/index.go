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
	"bytes"
	"strings"
	"github.com/jmhodges/levigo"
)

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

func DecodeDataKeys(keys []byte) []string {
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