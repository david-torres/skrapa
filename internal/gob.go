package internal

import (
	"bytes"
	"encoding/gob"
)

func gobEncode(s interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecodeStringSlice(s []byte) ([]string, error) {
	var dec []string
	buf := new(bytes.Buffer)
	buf.Write(s)
	decoder := gob.NewDecoder(buf)

	err := decoder.Decode(&dec)
	if err != nil {
		return nil, err
	}

	return dec, nil
}
