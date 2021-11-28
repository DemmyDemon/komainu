package storage

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/recoilme/sniper"
)

type Bytesliced interface {
	ToByteSlice() []byte
	FromByteSlice([]byte)
}

type SniperData struct {
	sniper *sniper.Store
}

var _singleton *SniperData

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

func BuildFinalKey(guildID discord.GuildID, collection string, key interface{}) []byte {
	return []byte(fmt.Sprintf("%s/%s/%v", guildID, collection, key))
}

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
		return fmt.Errorf("unsupported value type %T", rawValue)
	}
}

func (sd *SniperData) GetString(guild discord.GuildID, collection string, key interface{}) (bool, string, error) {
	if exist, value, err := sd.Get(guild, collection, key); err != nil {
		return false, "", err
	} else if !exist {
		return exist, "", nil
	} else {
		return exist, string(value), nil
	}
}

func (sd *SniperData) GetInt64(guild discord.GuildID, collection string, key interface{}) (bool, int64, error) {
	exist, value, err := sd.GetUint64(guild, collection, key)
	return exist, int64(value), err
}

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
