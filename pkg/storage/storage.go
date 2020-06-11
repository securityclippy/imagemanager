package storage

import "errors"

var ErrKeyNotFound = errors.New("key not found")


type Storage interface {
	Save(key string, src interface{}) error
	Load(key string, dst interface{}) error
}
