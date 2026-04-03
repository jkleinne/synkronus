package service

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReaderWithCleanup_CloseBothSucceed(t *testing.T) {
	inner := io.NopCloser(strings.NewReader("data"))
	r := &readerWithCleanup{ReadCloser: inner, cleanup: func() error { return nil }}

	if err := r.Close(); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestReaderWithCleanup_ReaderCloseError(t *testing.T) {
	r := &readerWithCleanup{
		ReadCloser: &errorCloser{readErr: errors.New("reader fail")},
		cleanup:    func() error { return nil },
	}

	err := r.Close()
	if err == nil || !strings.Contains(err.Error(), "reader fail") {
		t.Errorf("expected 'reader fail' error, got: %v", err)
	}
}

func TestReaderWithCleanup_CleanupError(t *testing.T) {
	inner := io.NopCloser(strings.NewReader("data"))
	r := &readerWithCleanup{
		ReadCloser: inner,
		cleanup:    func() error { return errors.New("cleanup fail") },
	}

	err := r.Close()
	if err == nil || !strings.Contains(err.Error(), "cleanup fail") {
		t.Errorf("expected 'cleanup fail' error, got: %v", err)
	}
}

func TestReaderWithCleanup_BothErrors(t *testing.T) {
	r := &readerWithCleanup{
		ReadCloser: &errorCloser{readErr: errors.New("reader fail")},
		cleanup:    func() error { return errors.New("cleanup fail") },
	}

	err := r.Close()
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(err.Error(), "reader fail") {
		t.Errorf("error should contain 'reader fail', got: %v", err)
	}
	if !strings.Contains(err.Error(), "cleanup fail") {
		t.Errorf("error should contain 'cleanup fail', got: %v", err)
	}
}

// errorCloser is an io.ReadCloser whose Close always returns an error.
type errorCloser struct {
	readErr error
}

func (e *errorCloser) Read(p []byte) (int, error) { return 0, io.EOF }
func (e *errorCloser) Close() error               { return e.readErr }
