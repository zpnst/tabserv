package database

import (
	"context"
	"sync"
)

type KVSValue struct {
	Value []byte
}

type DatabseInMemory struct {
	data     map[string]*KVSValue
	dataLock sync.RWMutex
}

func NewDatabseInMemory() *DatabseInMemory {
	return &DatabseInMemory{
		data: make(map[string]*KVSValue),
	}
}

func (db *DatabseInMemory) Put(_ context.Context, key string, value []byte) error {
	db.dataLock.Lock()
	defer db.dataLock.Unlock()
	val := make([]byte, len(value))
	copy(val, value)
	db.data[key] = &KVSValue{
		Value: val,
	}
	return nil
}

func (db *DatabseInMemory) Get(_ context.Context, key string) ([]byte, bool, error) {
	db.dataLock.RLock()
	defer db.dataLock.RUnlock()
	v, ok := db.data[key]
	if !ok {
		return nil, false, nil
	}
	val := v.Value

	out := make([]byte, len(val))
	copy(out, val)
	return out, true, nil
}

func (db *DatabseInMemory) Delete(_ context.Context, key string) (bool, error) {
	db.dataLock.Lock()
	defer db.dataLock.Unlock()
	if _, ok := db.data[key]; ok {
		delete(db.data, key)
		return true, nil
	}
	return false, nil
}

func (db *DatabseInMemory) All(ctx context.Context) (map[string][]byte, error) {
	db.dataLock.Lock()
	defer db.dataLock.Unlock()
	res := make(map[string][]byte, len(db.data))
	for k, val := range db.data {
		res[k] = val.Value
	}
	return res, nil
}
