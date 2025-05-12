package internal

import (
	"io"
)

func FixJsonStream(r io.Reader, w io.Writer) (bool, error) {
	lastChar := byte(0)
	buffer := make([]byte, 1024)
	changes := false

	for {
		n, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			return changes, err
		}

		hasChanges, processErr := processBuffer(buffer[:n], &lastChar, w)
		changes = changes || hasChanges
		if processErr != nil {
			return changes, processErr
		}

		if err == io.EOF {
			break
		}
	}
	return changes, nil
}

func processBuffer(buffer []byte, lastChar *byte, w io.Writer) (bool, error) {
	changes := false
	for _, char := range buffer {
		hasChanges, err := handleChar(char, lastChar, w)
		changes = changes || hasChanges
		if err != nil {
			return changes, err
		}
	}
	return changes, nil
}

func handleChar(char byte, lastChar *byte, w io.Writer) (bool, error) {
	changes := false
	if *lastChar == '}' && char == '{' {
		changes = true
		if _, err := w.Write([]byte{'\n', char}); err != nil {
			return changes, err
		}
	} else if (*lastChar == 0 || *lastChar == '\n') && char == '\n' {
		// Skip redundant newlines
		changes = true
	} else {
		if _, err := w.Write([]byte{char}); err != nil {
			return changes, err
		}
	}
	*lastChar = char
	return changes, nil
}
