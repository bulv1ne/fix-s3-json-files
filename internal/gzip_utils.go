package internal

import (
	"bytes"
	"compress/gzip"
)

func compress(input []byte) ([]byte, error) {
	buffer := &bytes.Buffer{}
	w := gzip.NewWriter(buffer)
	_, err := w.Write(input)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
