package prompt

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestConfirm_CorrectValue(t *testing.T) {
	input := strings.NewReader("my-bucket\n")
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	confirmed, err := p.Confirm("Delete bucket?", "my-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected confirmation to succeed")
	}
}

func TestConfirm_WrongValue(t *testing.T) {
	input := strings.NewReader("wrong-name\n")
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	confirmed, err := p.Confirm("Delete bucket?", "my-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected confirmation to fail with wrong value")
	}
}

func TestConfirm_EmptyExpectedValue(t *testing.T) {
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	_, err := p.Confirm("Delete?", "")
	if err == nil {
		t.Fatal("expected error for empty expected value")
	}
}

func TestConfirm_EOF(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	confirmed, err := p.Confirm("Delete?", "my-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected false on EOF")
	}
}

func TestConfirm_WhitespaceTrimming(t *testing.T) {
	input := strings.NewReader("  my-bucket  \n")
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	confirmed, err := p.Confirm("Delete?", "my-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected confirmation to succeed with trimmed whitespace")
	}
}

func TestConfirm_CaseSensitive(t *testing.T) {
	input := strings.NewReader("My-Bucket\n")
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	confirmed, err := p.Confirm("Delete?", "my-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected case-sensitive comparison to fail")
	}
}

func TestConfirm_ReadError(t *testing.T) {
	input := &errorReader{err: fmt.Errorf("device error")}
	output := &bytes.Buffer{}
	p := NewStandardPrompter(input, output)

	_, err := p.Confirm("Delete?", "my-bucket")
	if err == nil {
		t.Fatal("expected error for reader failure")
	}
	if !strings.Contains(err.Error(), "device error") {
		t.Errorf("expected wrapped error with 'device error', got: %v", err)
	}
}

// errorReader is a test helper that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}
