package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/kouleen/lets-encrypt/pkg/util"
)

type StoreRepository struct {
	cacheStore *util.ExpireMap
}

var storeRepository *StoreRepository
var storeOnce sync.Once

func getStoreRepository() *StoreRepository {
	storeOnce.Do(func() {
		storeRepository = &StoreRepository{
			util.GetCacheStore(),
		}
	})
	return storeRepository
}

func SetCacheAuth(key string, json string, duration time.Duration) error {
	if key == "" {
		return errors.New("key is empty")
	}
	getStoreRepository().cacheStore.Set(key, json, duration)
	return nil
}

func GetCacheAuth(key string) (string, error) {
	if key == "" {
		return "", errors.New("key is empty")
	}
	return getStoreRepository().cacheStore.Get(key), nil
}

func DelCacheAuth(key string) error {
	getStoreRepository().cacheStore.Delete(key)
	return nil
}

func TtlCacheAuth(key string) time.Duration {
	if key == "" {
		return -2
	}
	return getStoreRepository().cacheStore.Ttl(key)
}
