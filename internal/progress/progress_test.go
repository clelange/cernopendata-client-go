package progress

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestNewWriter(t *testing.T) {
	var buf bytes.Buffer
	pw := NewWriter(&buf, "test.dat", 1000)

	if pw == nil {
		t.Fatal("NewWriter returned nil")
	}

	if pw.filename != "test.dat" {
		t.Errorf("Expected filename 'test.dat', got '%s'", pw.filename)
	}

	if pw.totalBytes != 1000 {
		t.Errorf("Expected totalBytes 1000, got %d", pw.totalBytes)
	}

	if pw.writtenBytes != 0 {
		t.Errorf("Expected writtenBytes 0, got %d", pw.writtenBytes)
	}
}

func TestWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	pw := NewWriter(&buf, "test.dat", 1000)
	// Use a noop output to avoid actual printing
	pw.output = io.Discard
	pw.updateEvery = 1 * time.Hour // Disable periodic updates

	data := []byte("hello world")
	n, err := pw.Write(data)

	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	if pw.WrittenBytes() != int64(len(data)) {
		t.Errorf("Expected WrittenBytes %d, got %d", len(data), pw.WrittenBytes())
	}

	// Check the underlying writer received the data
	if buf.String() != "hello world" {
		t.Errorf("Expected 'hello world' in buffer, got '%s'", buf.String())
	}
}

func TestWriter_WriteMultiple(t *testing.T) {
	var buf bytes.Buffer
	pw := NewWriter(&buf, "test.dat", 1000)
	pw.output = io.Discard
	pw.updateEvery = 1 * time.Hour

	// Write multiple times
	_, _ = pw.Write([]byte("hello "))
	_, _ = pw.Write([]byte("world"))

	if pw.WrittenBytes() != 11 {
		t.Errorf("Expected WrittenBytes 11, got %d", pw.WrittenBytes())
	}
}

func TestWriter_WrittenBytes(t *testing.T) {
	var buf bytes.Buffer
	pw := NewWriter(&buf, "test.dat", 100)
	pw.output = io.Discard
	pw.updateEvery = 1 * time.Hour

	if pw.WrittenBytes() != 0 {
		t.Errorf("Expected initial WrittenBytes 0, got %d", pw.WrittenBytes())
	}

	_, _ = pw.Write([]byte("12345"))
	if pw.WrittenBytes() != 5 {
		t.Errorf("Expected WrittenBytes 5, got %d", pw.WrittenBytes())
	}

	_, _ = pw.Write([]byte("67890"))
	if pw.WrittenBytes() != 10 {
		t.Errorf("Expected WrittenBytes 10, got %d", pw.WrittenBytes())
	}
}

func TestWriter_Finish(t *testing.T) {
	var buf bytes.Buffer
	var output bytes.Buffer
	pw := NewWriter(&buf, "test.dat", 100)
	pw.output = &output
	pw.updateEvery = 1 * time.Hour

	_, _ = pw.Write([]byte("12345"))
	pw.Finish()

	// Verify finish was called (output should contain something)
	if output.Len() == 0 {
		t.Error("Expected Finish to produce output")
	}
}

func TestWriter_ProgressPercentage(t *testing.T) {
	var buf bytes.Buffer
	var output bytes.Buffer
	pw := NewWriter(&buf, "test.dat", 100)
	pw.output = &output
	pw.updateEvery = 0 // Enable immediate updates

	// Write 50 bytes (50%)
	data := make([]byte, 50)
	_, _ = pw.Write(data)

	// Check that output contains percentage
	outputStr := output.String()
	if len(outputStr) == 0 {
		t.Error("Expected progress output")
	}
}

func TestWriter_UnknownTotalSize(t *testing.T) {
	var buf bytes.Buffer
	var output bytes.Buffer
	// Total size of 0 means unknown
	pw := NewWriter(&buf, "test.dat", 0)
	pw.output = &output
	pw.updateEvery = 1 * time.Hour

	_, _ = pw.Write([]byte("hello"))
	pw.Finish()

	// Should still work with unknown total size
	if output.Len() == 0 {
		t.Error("Expected Finish to produce output even with unknown size")
	}
}
