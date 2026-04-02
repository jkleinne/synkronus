package prompt

import (
	"bytes"
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
