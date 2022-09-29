package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	bolt "go.etcd.io/bbolt"
)

type komainuBolt struct {
	bolt *bolt.DB
}

func OpenKomainuBolt(path string) (*komainuBolt, error) {
	newBolt, err := bolt.Open(path, 0660, nil)
	if err != nil {
		return nil, err
	}
	return &komainuBolt{
		newBolt,
	}, nil
}

func (kb *komainuBolt) createBucket(transaction *bolt.Tx, guild []byte, collection []byte) (bucket *bolt.Bucket, err error) {
	guildBucket, err := transaction.CreateBucketIfNotExists(guild)
	if err != nil {
		return nil, fmt.Errorf("bolt store failed to create guild bucket: %w", err)
	}
	bucket, err = guildBucket.CreateBucketIfNotExists(collection)
	if err != nil {
		return nil, fmt.Errorf("bolt store failed to create collection bucket: %w", err)
	}
	return
}

func (kb *komainuBolt) getBucket(transaction *bolt.Tx, guild []byte, collection []byte) (bucket *bolt.Bucket) {
	guildBucket := transaction.Bucket(guild)
	if guildBucket == nil {
		return nil
	}
	return guildBucket.Bucket(collection)
}

func (kb *komainuBolt) store(guild []byte, collection []byte, key []byte, value []byte) (err error) {
	return kb.bolt.Update(func(tx *bolt.Tx) (err error) {
		if tx == nil {
			return errors.New("storage failed to open Update transaction")
		}
		bucket, err := kb.createBucket(tx, guild, collection)
		if err != nil {
			return
		}
		err = bucket.Put(key, value)
		if err != nil {
			err = fmt.Errorf("bolt store failed to Put value: %w", err)
		}
		return
	})
}

func (kb *komainuBolt) retrieve(guild []byte, collection []byte, key []byte) (found bool, value []byte, err error) {
	err = kb.bolt.View(func(tx *bolt.Tx) (err error) {
		if tx == nil {
			return errors.New("storage failed to open View transaction")
		}
		bucket := kb.getBucket(tx, guild, collection)
		if bucket == nil {
			return
		}
		got := bucket.Get(key)
		if got != nil {
			found = true
			value = make([]byte, len(got))
			copy(value, got)
		}
		return
	})
	return
}

func (kb *komainuBolt) remove(guild []byte, collection []byte, key []byte) (err error) {
	err = kb.bolt.Update(func(tx *bolt.Tx) (err error) {
		if tx == nil {
			return errors.New("storage failed to open Delete transaction")
		}
		bucket := kb.getBucket(tx, guild, collection)
		if bucket == nil {
			return
		}
		err = bucket.Delete(key)
		if err != nil {
			return fmt.Errorf("storage failed to delete: %w", err)
		}
		return
	})
	return
}

func (kb *komainuBolt) keys(guild []byte, collection []byte) (keys []string, err error) {
	err = kb.bolt.View(func(tx *bolt.Tx) (err error) {
		bucket := kb.getBucket(tx, guild, collection)
		if bucket == nil {
			return
		}
		bucket.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
		return
	})
	return
}

func (kb *komainuBolt) key(raw any) []byte {
	return []byte(fmt.Sprintf("%v", raw))
}

func (kb *komainuBolt) Set(guildID discord.GuildID, collection string, key any, value any) (err error) {
	guildb := []byte(guildID.String())
	collectionb := []byte(collection)
	keyb := kb.key(key)
	var inputBuffer bytes.Buffer
	if err := gob.NewEncoder(&inputBuffer).Encode(value); err != nil {
		return fmt.Errorf("unable to encode raw value %v as gob for Set: %w", value, err)
	}
	return kb.store(guildb, collectionb, keyb, inputBuffer.Bytes())
}

func (kb *komainuBolt) Get(guildID discord.GuildID, collection string, key any, out any) (found bool, err error) {
	guildb := []byte(guildID.String())
	collectionb := []byte(collection)
	keyb := kb.key(key)
	found, raw, err := kb.retrieve(guildb, collectionb, keyb)
	if err != nil || !found {
		return
	}
	var outputBuffer bytes.Buffer
	outputBuffer.Write(raw)
	err = gob.NewDecoder(&outputBuffer).Decode(out)
	return
}

func (kb *komainuBolt) Delete(guildID discord.GuildID, collection string, key any) (err error) {
	guildb := []byte(guildID.String())
	collectionb := []byte(collection)
	keyb := kb.key(key)
	return kb.remove(guildb, collectionb, keyb)
}

func (kb *komainuBolt) Keys(guildID discord.GuildID, collection string) (keys []string, err error) {
	guildb := []byte(guildID.String())
	collectionb := []byte(collection)
	return kb.keys(guildb, collectionb)
}

func (kb *komainuBolt) Close() error {
	return kb.bolt.Close()
}
