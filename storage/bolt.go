package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"math"

	"github.com/diamondburned/arikawa/v3/discord"
	bolt "go.etcd.io/bbolt"
)

type boltData struct {
	bolt *bolt.DB
}

// OpenBolt opens the bolt database at the given path for handling.
func OpenBolt(path string) (*boltData, error) {
	newBolt, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &boltData{
		bolt: newBolt,
	}, nil
}

// Open opens the bolt database at the given path for handling.
func (b *boltData) Open(path string) error {
	newBolt, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	b.bolt = newBolt
	return nil
}

// Close closes the bolt database behind the scenes.
func (b *boltData) Close() error {
	return b.bolt.Close()
}

// store is the unexported backing store routine to actually write to bolt
func (b *boltData) store(bucketName []byte, key []byte, value []byte) (err error) {
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

// bucketName returns a byte array based on the guild and collection given, in byte slice form, suitable for use as a bucket name.
func (b *boltData) bucketName(guild discord.GuildID, collection string) (bucketName []byte) {
	return []byte(fmt.Sprintf("%s/%s", guild, collection))
}

// keyName takes almost any random thing and tries to make a string, and then a byte slice out of it, returning that byte slice.
// Used to make a byte slice suitable as a key in bolt
func (b *boltData) keyName(key interface{}) (keyName []byte) {
	return []byte(fmt.Sprintf("%v", key))
}

// Set takes a guild, collection and key, along with a raw value, and tries to stuff that value into bolt in a somewhat sane manner.
func (b *boltData) Set(guild discord.GuildID, collection string, key interface{}, rawValue interface{}) (err error) {
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
	case float64:
		buf := make([]byte, 8)
		bits := math.Float64bits(value)
		binary.LittleEndian.PutUint64(buf, bits)
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

// Get takes a guild, collection and key, and tries to look up the value stored in bolt under that key.
func (b *boltData) Get(guild discord.GuildID, collection string, key interface{}) (exist bool, value []byte, err error) {
	bucketName := b.bucketName(guild, collection)
	keyName := b.keyName(key)

	err = b.bolt.View(func(tx *bolt.Tx) error {
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
			exist = true
		}
		return nil
	})

	return exist, value, err
}

// GetObject attempts to decode whatever is stored under the given key into the target reference.
// Best of luck!
func (b *boltData) GetObject(guild discord.GuildID, collection string, key interface{}, target interface{}) (exist bool, err error) {
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		var outputBuffer bytes.Buffer
		outputBuffer.Write(raw)
		err = gob.NewDecoder(&outputBuffer).Decode(target)
	}
	return exist, err
}

// GetString looks up the data under the given guild, collection and key, assumes it's a string and returns if the key exists, the found string and any error encountered.
func (b *boltData) GetString(guild discord.GuildID, collection string, key interface{}) (exist bool, value string, err error) {
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		value = string(raw)
	}
	return exist, value, err
}

// GetInt64 looks up the data under the given guild, collection and key, assumes it's an int64 and returns if the key exists, the found int64 and any error encountered.
func (b *boltData) GetInt64(guild discord.GuildID, collection string, key interface{}) (exist bool, value int64, err error) {
	exist, raw, err := b.GetUint64(guild, collection, key)
	return exist, int64(raw), err
}

// GetInt64 looks up the data under the given guild, collection and key, assumes it's a uint64 and returns if the key exists, the found uint64 and any error encountered.
func (b *boltData) GetUint64(guild discord.GuildID, collection string, key interface{}) (exist bool, value uint64, err error) {
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		value = binary.LittleEndian.Uint64(raw)
	}
	return exist, value, err
}

// GetFloat64 looks up the data under the given guild, collection and key, assumes it's a float64 and returns if the key exists, the found float64 and any error encountered.
func (b *boltData) GetFloat64(guild discord.GuildID, collection string, key interface{}) (exist bool, value float64, err error) {
	exist, raw, err := b.Get(guild, collection, key)

	if err == nil && exist {
		asUint := binary.LittleEndian.Uint64(raw)
		value = math.Float64frombits(asUint)
	}
	return exist, value, err
}

// Delete checks if it is possible to delete the value under the given guild, collection and key (as in "The Bucket Exists"), and attempts to do so.
// It returns if a delete attempt was made and any error encountered.
// If the last key in a bucket is removed, it also removes the bucket.
func (b *boltData) Delete(guild discord.GuildID, collection string, key interface{}) (wasDeleted bool, err error) {
	bucketName := b.bucketName(guild, collection)
	keyName := b.keyName(key)

	err = b.bolt.Update(func(tx *bolt.Tx) error {
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

		if bucket.Stats().KeyN == 0 { // This is a lot of processing for just one comparison. It's fine, this is dwarved by all the Discord HTTP stuff.
			if err := tx.DeleteBucket(bucketName); err != nil {
				return fmt.Errorf("deleted key was last in bucket, but deleting bucket failed: %w", err)
			}
		}
		return nil
	})

	return wasDeleted, err
}

// Keys returns a slice of strings representing the keys in the given guild and collection.
// Note that if the keys originally used were not strings, this will be garbage data.
// If the bucket does not exists, or has no keys, a zero-length slice is returned.
func (b *boltData) Keys(guild discord.GuildID, collection string) (keys []string, err error) {
	bucketName := b.bucketName(guild, collection)

	err = b.bolt.View(func(tx *bolt.Tx) error {
		if tx == nil {
			return errors.New("could not get Keys from bolt: failed to open transaction")
		}
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil
		}
		bucket.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
		return nil
	})
	return keys, err
}
