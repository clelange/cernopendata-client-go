package printer

import (
	"fmt"
	"os"
)

type MessageType int

const (
	Info MessageType = iota
	Note
	Progress
	Error
)

func DisplayMessage(msgType MessageType, message string) {
	switch msgType {
	case Info:
		_, _ = fmt.Fprintf(os.Stdout, "%s%s\n", "==> ", message)
	case Note, Progress:
		_, _ = fmt.Fprintf(os.Stdout, "  -> %s\n", message)
	case Error:
		_, _ = fmt.Fprintf(os.Stderr, "==> ERROR: %s\n", message)
	}
}

func DisplayOutput(message string) {
	_, _ = fmt.Fprintln(os.Stdout, message)
}
