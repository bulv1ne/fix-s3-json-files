package internal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFixJsonStream(t *testing.T) {
	reader := strings.NewReader("{\"foo\":\"bar\"}{\"foo\":\"baz\"}\n\n{\"foo\":\"qux\"}")
	expected := "{\"foo\":\"bar\"}\n{\"foo\":\"baz\"}\n{\"foo\":\"qux\"}"
	writer := &strings.Builder{}
	changes, err := FixJsonStream(reader, writer)
	assert.NoError(t, err)
	assert.True(t, changes)
	result := writer.String()
	assert.Equal(t, expected, result)
}

func TestFixJsonStreamNoChanges(t *testing.T) {
	expected := "{\"foo\":\"bar\"}\n{\"foo\":\"baz\"}\n{\"foo\":\"qux\"}"
	reader := strings.NewReader(expected)
	writer := &strings.Builder{}

	changes, err := FixJsonStream(reader, writer)
	assert.NoError(t, err)
	assert.False(t, changes)
	result := writer.String()
	assert.Equal(t, expected, result)
}

const expectedSize = 10000

func BenchmarkFixJsonStream(b *testing.B) {
	expected := "{\"foo\":\"bar\"}\n{\"foo\":\"baz\"}\n{\"foo\":\"qux\"}"
	bigExpected := &strings.Builder{}
	for range expectedSize {
		bigExpected.WriteString(expected)
	}
	compressedData, _ := compress([]byte(bigExpected.String()))
	b.ResetTimer()
	for b.Loop() {
		reader := bytes.NewReader(compressedData)
		reader2, _ := gzip.NewReader(reader)

		writer := &strings.Builder{}
		writer2 := gzip.NewWriter(writer)
		_, _ = FixJsonStream(reader2, writer2)
		_ = reader2.Close()
		_ = writer2.Close()
	}
}

func BenchmarkFixJsonStreamBuffer(b *testing.B) {
	expected := "{\"foo\":\"bar\"}\n{\"foo\":\"baz\"}\n{\"foo\":\"qux\"}"
	bigExpected := &strings.Builder{}
	for range expectedSize {
		bigExpected.WriteString(expected)
	}
	compressedData, _ := compress([]byte(bigExpected.String()))
	b.ResetTimer()
	for b.Loop() {
		reader := bytes.NewReader(compressedData)
		reader2, _ := gzip.NewReader(reader)

		writer := &strings.Builder{}
		writer2 := gzip.NewWriter(writer)
		writer3 := bufio.NewWriter(writer2)
		_, _ = FixJsonStream(reader2, writer3)
		_ = reader2.Close()
		_ = writer3.Flush()
		_ = writer2.Close()
	}
}
