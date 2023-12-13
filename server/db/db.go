package db

import (
	"log"

	"github.com/potproject/uchinoko/envgen"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB

func Start() {
	var err error
	db, err = leveldb.OpenFile(envgen.Get().DB_FILE_PATH(), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func put(key []byte, value []byte) error {
	return db.Put(key, value, nil)
}

func get(key []byte) ([]byte, error) {
	return db.Get(key, nil)
}

type ListAllResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ListAll() ([]ListAllResponse, error) {
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	var response []ListAllResponse
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		response = append(response, ListAllResponse{
			Key:   string(key),
			Value: string(value),
		})
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return response, nil
}
