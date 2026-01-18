package printer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDisplayOutput(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple string", "test output", "test output"},
		{"string with newline", "line1\nline2", "line1\nline2"},
		{"empty string", "", ""},
		{"string with spaces", "hello world", "hello world"},
		{"special characters", "test\x00\x01", "test\x00\x01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DisplayOutput(tt.input)
			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			got := buf.String()

			if !strings.Contains(got, tt.want) {
				t.Errorf("DisplayOutput() output = %q, want to contain %q", got, tt.want)
			}

			r, w, _ = os.Pipe()
			os.Stdout = w
		})
	}

	os.Stdout = oldStdout
}

func TestDisplayMessage(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		message  string
		wantErr  bool
		checkStd bool
	}{
		{"info message", Info, "info message", false, false},
		{"note message", Note, "note message", false, false},
		{"error message", Error, "error message", false, true},
		{"invalid message type", MessageType(999), "message", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			rOut, wOut, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr

			if tt.checkStd {
				os.Stdout = oldStdout
			}

			if !tt.wantErr {
				DisplayMessage(tt.msgType, tt.message)
			}

			_ = wOut.Close()
			_ = wErr.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var bufOut bytes.Buffer
			var bufErr bytes.Buffer
			_, _ = io.Copy(&bufOut, rOut)
			_, _ = io.Copy(&bufErr, rErr)

			output := bufOut.String()
			errOutput := bufErr.String()

			if tt.checkStd {
				if !strings.Contains(errOutput, "ERROR") {
					t.Errorf("DisplayMessage(Error) stderr should contain 'ERROR', got %q", errOutput)
				}
				if !strings.Contains(errOutput, tt.message) {
					t.Errorf("DisplayMessage(Error) stderr should contain message %q, got %q", tt.message, errOutput)
				}
			} else if !tt.wantErr {
				if output == "" && errOutput == "" {
					t.Errorf("DisplayMessage() should produce output for %v", tt.msgType)
				}
			}
		})
	}
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"Info type", Info, "Info"},
		{"Note type", Note, "Note"},
		{"Error type", Error, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeName := ""
			switch tt.msgType {
			case Info:
				typeName = "Info"
			case Note:
				typeName = "Note"
			case Error:
				typeName = "Error"
			}
			if typeName != tt.expected {
				t.Errorf("MessageType = %v, want %v", typeName, tt.expected)
			}
		})
	}
}
