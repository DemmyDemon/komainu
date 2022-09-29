package storage

import (
	"os"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func GetKVS(path string) (KeyValueStore, error) {
	return OpenKomainuBolt(path)
}

const filename = "test_storage_file"
const col = "test collection"

var testGuild = discord.GuildID(211575243083350016)

/*
	Yes, I'm aware the test coverage is atrocious.
	I'm testing most of it by hand anyway, and I'm currently writing a boatload of tests at work, so sue me.
	This is better testing than *any* of the other stuff!
*/

func TestStorageInt(t *testing.T) {
	kvs, err := GetKVS(filename)
	if err != nil {
		t.Errorf("Could not open test file: %s", err)
	}
	t.Cleanup(func() {
		kvs.Close()
		os.Remove(filename)
	})

	key := "some arbituary key"
	input := 12345
	if err := kvs.Set(testGuild, col, key, input); err != nil {
		t.Errorf("Could not set test input value: %v", err)
		return
	}

	var output int
	found, err := kvs.Get(testGuild, col, key, &output)
	if err != nil {
		t.Errorf("Could not retrieve value: %v", err)
	}
	if !found {
		t.Error("Value was not found when trying to read it back!")
	}
	if input != output {
		t.Errorf("Expected %d, Got %d", input, output)
	}
}

func TestStorageString(t *testing.T) {
	kvs, err := GetKVS(filename)
	if err != nil {
		t.Errorf("Could not open test file: %s", err)
	}
	t.Cleanup(func() {
		kvs.Close()
		os.Remove(filename)
	})

	key := "some arbituary key"
	input := "some arbituary string"
	if err := kvs.Set(testGuild, col, key, input); err != nil {
		t.Errorf("Could not set test input value: %v", err)
		return
	}

	var output string
	found, err := kvs.Get(testGuild, col, key, &output)
	if err != nil {
		t.Errorf("Could not retrieve value: %v", err)
	}
	if !found {
		t.Error("Value was not found when trying to read it back!")
	}
	if input != output {
		t.Errorf("Expected %s, Got %s", input, output)
	}
}

func TestStorageObject(t *testing.T) {
	kvs, err := GetKVS(filename)
	if err != nil {
		t.Errorf("Could not open test file: %s", err)
	}
	t.Cleanup(func() {
		kvs.Close()
		os.Remove(filename)
	})

	type SomeTestStruct struct {
		First  string
		Second int
	}

	key := "some arbituary key"
	input := SomeTestStruct{"foo", 123}
	if err := kvs.Set(testGuild, col, key, input); err != nil {
		t.Errorf("Could not set test input value: %v", err)
		return
	}

	var output SomeTestStruct
	found, err := kvs.Get(testGuild, col, key, &output)
	if err != nil {
		t.Errorf("Could not retrieve value: %v", err)
	}
	if !found {
		t.Error("Value was not found when trying to read it back!")
	}
	if input.First != output.First {
		t.Errorf("First field: Expected %s, Got %s", input.First, output.First)
	}
	if input.Second != output.Second {
		t.Errorf("Second field: Expected %d, Got %d", input.Second, output.Second)
	}
}

func TestMissingValue(t *testing.T) {
	kvs, err := GetKVS(filename)
	if err != nil {
		t.Errorf("Could not open test file: %s", err)
	}
	defer func() {
		kvs.Close()
		os.Remove(filename)
	}()
}
