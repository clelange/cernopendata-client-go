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
		fmt.Fprintf(os.Stdout, "%s%s\n", "==> ", message)
	case Note, Progress:
		fmt.Fprintf(os.Stdout, "  -> %s\n", message)
	case Error:
		fmt.Fprintf(os.Stderr, "==> ERROR: %s\n", message)
	}
}

func DisplayOutput(message string) {
	fmt.Fprintln(os.Stdout, message)
}
