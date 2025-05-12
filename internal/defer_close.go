package internal

import (
	"fmt"
	"io"
)

func CloseAndLogOnError(closer io.Closer, msg ...string) {
	if err := closer.Close(); err != nil {
		if len(msg) > 0 {
			fmt.Println("Error closing file", msg[0], err)
		} else {
			fmt.Println("Error closing file", err)
		}
	}
}
