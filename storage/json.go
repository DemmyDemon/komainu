package storage

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
)

type JSONStorable interface {
	Path() string
	Load() error
	Save() error
}

// JSONFileExists checks if the given JSONStorable's .Path() points to a file that already exists.
func JSONFileExists(storable JSONStorable) (bool, error) {
	if _, err := os.Stat(storable.Path()); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// MarshalJSON is a function that marshals the object into an io.Reader.
func MarshalJSON(storable JSONStorable) (io.Reader, error) {

	resultingBytes, err := json.MarshalIndent(storable, "", "\t")
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(resultingBytes), nil
}

// UnmarshalJSON is a function that unmarshals the data from the reader into the specified value.
func UnmarshalJSON(r io.Reader, storable JSONStorable) error {
	return json.NewDecoder(r).Decode(storable)
}

// LoadJSON loads the given JSONStorable.
// Use os.IsNotExist() to see if the returned error is due to the file being missing.
func LoadJSON(storable JSONStorable) error {

	fileHandle, err := os.Open(storable.Path())
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	return UnmarshalJSON(fileHandle, storable)
}

// SaveJSON saves a representation of the given JSONStorable
func SaveJSON(storable JSONStorable) error {
	if err := os.MkdirAll(filepath.Dir(storable.Path()), 0770); err != nil {
		return err
	}
	fileHandle, err := os.Create(storable.Path())
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	reader, err := MarshalJSON(storable)
	if err != nil {
		return err
	}
	size, err := io.Copy(fileHandle, reader)
	log.Printf("Saved %d bytes to %s\n", size, storable.Path())
	return err
}
