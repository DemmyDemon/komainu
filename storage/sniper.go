package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/recoilme/sniper"
)

type SniperData struct {
	sniper *sniper.Store
}

var _singleton *SniperData

// Sniper returns (or creates) the Sniper instance for storage needs.
func Sniper() *SniperData {
	if _singleton != nil {
		return _singleton
	}
	if newSniper, err := sniper.Open(sniper.Dir("data/sniper")); err != nil {
		log.Fatalln("Failed to open sniper data source:", err)
	} else {
		_singleton = &SniperData{
			sniper: newSniper,
		}
	}
	return _singleton
}

// BuildFinalKey naively concatinates together a bunch of stuff to make an appropriate key to store data under.
func BuildFinalKey(guildID discord.GuildID, collection string, key interface{}) []byte {
	return []byte(fmt.Sprintf("%s/%s/%v", guildID, collection, key))
}

// Get gets []byte data from Sniper.
func (sd *SniperData) Get(guild discord.GuildID, collection string, key interface{}) (bool, []byte, error) {
	finalKey := BuildFinalKey(guild, collection, key)
	if val, err := sd.sniper.Get(finalKey); err != nil {
		if err == sniper.ErrNotFound {
			return false, nil, nil
		} else {
			return false, nil, err
		}
	} else {
		return true, val, nil
	}
}

// Set smushes data into Sniper using a rubber mallet and lube.
func (sd *SniperData) Set(guild discord.GuildID, collection string, key interface{}, rawValue interface{}) error {
	finalKey := BuildFinalKey(guild, collection, key)
	switch finalValue := rawValue.(type) {
	case []byte:
		return sd.sniper.Set(finalKey, finalValue, 0)
	case int64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(finalValue))
		return sd.sniper.Set(finalKey, b, 0)
	case uint64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, finalValue)
		log.Println(string(b))
		return sd.sniper.Set(finalKey, b, 0)
	case string:
		return sd.sniper.Set(finalKey, []byte(finalValue), 0)
	default:
		var inputBuffer bytes.Buffer
		if err := gob.NewEncoder(&inputBuffer).Encode(rawValue); err != nil {
			return err
		}
		return sd.sniper.Set(finalKey, inputBuffer.Bytes(), 0)
	}
}

// GetObject attempts to decode whatever is stored under the given key into the target reference.
// Best of luck.
func (sd *SniperData) GetObject(guild discord.GuildID, collection string, key interface{}, target interface{}) (bool, error) {
	exists, raw, err := sd.Get(guild, collection, key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	var outputBuffer bytes.Buffer
	outputBuffer.Write(raw)
	err = gob.NewDecoder(&outputBuffer).Decode(target)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetString gets data from Sniper, assumes it's a string, and gives it to you.
func (sd *SniperData) GetString(guild discord.GuildID, collection string, key interface{}) (bool, string, error) {
	if exist, value, err := sd.Get(guild, collection, key); err != nil {
		return false, "", err
	} else if !exist {
		return exist, "", nil
	} else {
		return exist, string(value), nil
	}
}

// GetInt64 gets data from Sniper, forces it to pretend to be an int64, and hands it to you.
func (sd *SniperData) GetInt64(guild discord.GuildID, collection string, key interface{}) (bool, int64, error) {
	exist, value, err := sd.GetUint64(guild, collection, key)
	return exist, int64(value), err
}

// GetUint64 gets data from Sniper, assumes it's a little endian uint64 and passes it to you as such.
func (sd *SniperData) GetUint64(guild discord.GuildID, collection string, key interface{}) (bool, uint64, error) {
	if exist, value, err := sd.Get(guild, collection, key); err != nil {
		return false, 0, err
	} else if !exist {
		return exist, 0, nil
	} else {
		finalValue := binary.LittleEndian.Uint64(value)
		return exist, finalValue, nil
	}
}

func (sd *SniperData) Delete(guild discord.GuildID, collection string, key interface{}) (bool, error) {
	finalKey := BuildFinalKey(guild, collection, key)
	return sd.sniper.Delete(finalKey)
}
