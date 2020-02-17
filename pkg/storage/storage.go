package storage



type Storage interface {
	Save(key string, src interface{})
	Load(key string, dst interface{})
}
