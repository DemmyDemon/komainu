package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	bolt "go.etcd.io/bbolt"
)

type boltData struct {
	bolt *bolt.DB
}

func OpenBolt(path string) (*boltData, error) {
	newBolt, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &boltData{
		bolt: newBolt,
	}, nil
}

func (b *boltData) Open(path string) error {
	newBolt, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	b.bolt = newBolt
	return nil
}

func (b *boltData) Close() error {
	return b.bolt.Close()
}

func (b *boltData) store(bucketName []byte, key []byte, value []byte) error {
	return b.bolt.Update(func(tx *bolt.Tx) error {
		if tx == nil {
			return errors.New("could not store to bolt: failed to open transaction")
		}
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("store failed to create bucket %s: %w", bucketName, err)
		}
		if err := b.Put(key, value); err != nil {
			return fmt.Errorf("store failed to Put in bucket %s: %w", bucketName, err)
		}
		return nil
	})
}

func (b *boltData) bucketName(guild discord.GuildID, collection string) []byte {
	return []byte(fmt.Sprintf("%s/%s", guild, collection))
}

func (b *boltData) keyName(key interface{}) []byte {
	return []byte(fmt.Sprintf("%v", key))
}

func (b *boltData) Set(guild discord.GuildID, collection string, key interface{}, rawValue interface{}) error {
	bucketName := b.bucketName(guild, collection)
	keyName := b.keyName(key)

	switch value := rawValue.(type) {
	case []byte:
		return b.store(bucketName, keyName, value)
	case int64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(value))
		return b.store(bucketName, keyName, buf)
	case uint64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, value)
		return b.store(bucketName, keyName, buf)
	case string:
		return b.store(bucketName, keyName, []byte(value))
	default:
		var inputBuffer bytes.Buffer
		if err := gob.NewEncoder(&inputBuffer).Encode(rawValue); err != nil {
			return fmt.Errorf("unable to encode raw value %v as gob for Set: %w", rawValue, err)
		}
		return b.store(bucketName, keyName, inputBuffer.Bytes())
	}
}

func (b *boltData) Get(guild discord.GuildID, collection string, key interface{}) (bool, []byte, error) {
	bucketName := b.bucketName(guild, collection)
	keyName := b.keyName(key)

	exists := false
	var value []byte

	err := b.bolt.View(func(tx *bolt.Tx) error {
		if tx == nil {
			return errors.New("could not Get from bolt: failed to open transaction")
		}
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil // This is not an error, it simply does not exist.
		}
		got := bucket.Get(keyName)
		if got != nil {
			value = got // Implicit copy
			exists = true
		}
		return nil
	})

	return exists, value, err
}

func (b *boltData) GetObject(guild discord.GuildID, collection string, key interface{}, target interface{}) (bool, error) {
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		var outputBuffer bytes.Buffer
		outputBuffer.Write(raw)
		err = gob.NewDecoder(&outputBuffer).Decode(target)
	}
	return exist, err
}

func (b *boltData) GetString(guild discord.GuildID, collection string, key interface{}) (bool, string, error) {
	value := ""
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		value = string(raw)
	}
	return exist, value, err
}

func (b *boltData) GetInt64(guild discord.GuildID, collection string, key interface{}) (bool, int64, error) {
	exist, value, err := b.GetUint64(guild, collection, key)
	return exist, int64(value), err
}

func (b *boltData) GetUint64(guild discord.GuildID, collection string, key interface{}) (bool, uint64, error) {
	value := uint64(0)
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		value = binary.LittleEndian.Uint64(raw)
	}
	return exist, value, err

}

func (b *boltData) Delete(guild discord.GuildID, collection string, key interface{}) (bool, error) {
	bucketName := b.bucketName(guild, collection)
	keyName := b.keyName(key)

	wasDeleted := false

	err := b.bolt.Update(func(tx *bolt.Tx) error {
		if tx == nil {
			return errors.New("could not delete from bolt: failed to open transaction")
		}
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil // not an error, it was just not deleted because it didn't exist.
		}
		if err := bucket.Delete(keyName); err != nil {
			return err
		}
		wasDeleted = true // NOTE: This could be a lie, but we're not going to Get the value just to check if it existed before.

		if bucket.Stats().KeyN == 0 { // FIXME: This is a lot of processing for just one comparison. Just leave the empty bucket?
			if err := tx.DeleteBucket(bucketName); err != nil {
				return fmt.Errorf("deleted key was last in bucket, but deleting bucket failed: %w", err)
			}
		}
		return nil
	})

	return wasDeleted, err
}
