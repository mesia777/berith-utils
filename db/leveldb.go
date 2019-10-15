package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
)

type Database struct {
	db   *leveldb.DB
	path string
}

// NewDatabase returns a new db
func NewDatabase(path string, o *opt.Options) (*Database, error) {
	if path == "" {
		return nil, errors.New("invalid path")
	}

	var opts opt.Options
	if o != nil {
		opts = *o
	}

	db, err := leveldb.OpenFile(path, &opts)
	if errors.IsCorrupted(err) {
		db, err = leveldb.RecoverFile(path, &opts)
		if err != nil {
			return nil, err
		}
	}
	log.Println("Open local database: ", path)

	return &Database{
		db:   db,
		path: path,
	}, nil
}

// Has returns true if its present in the store, otherwise false.
func (db *Database) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

// Put inserts the given value into the store.
func (db *Database) Put(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

// Get returns a value given key.
func (db *Database) Get(key []byte) ([]byte, error) {
	val, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// NewIterator returns the entire key space.
func (db *Database) NewIterator() iterator.Iterator {
	return db.db.NewIterator(new(util.Range), nil)
}

// NewIteratorWithStart returns the key space given subset start.
func (db *Database) NewIteratorWithStart(start []byte) iterator.Iterator {
	return db.db.NewIterator(&util.Range{Start: start}, nil)
}

// NewIteratorWithPrefix returns the key space given prefix.
func (db *Database) NewIteratorWithPrefix(p []byte) iterator.Iterator {
	return db.db.NewIterator(util.BytesPrefix(p), nil)
}

// Delete removes the key from the store.
func (db *Database) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

// Path returns the path to db directory
func (db *Database) Path() string {
	return db.path
}

func (db *Database) Close() {
	_ = db.db.Close()
}
