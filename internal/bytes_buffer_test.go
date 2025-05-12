package internal

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestBytesBuffer(t *testing.T) {
	output := &bytes.Buffer{}
	output.WriteString("hello world")

	result, _ := io.ReadAll(output)
	if string(result) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(result))
	}
}

func TestBytesBuffer2(t *testing.T) {
	output := &bytes.Buffer{}
	output.WriteString("hello world")

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			result := output.Bytes()
			if string(result) != "hello world" {
				t.Errorf("Expected 'hello world', got '%s'", string(result))
			}
		})
	}
}
