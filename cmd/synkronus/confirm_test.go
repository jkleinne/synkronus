package main

import (
	"errors"
	"strings"
	"testing"
	"synkronus/internal/ui/prompt"
)

// mockPrompter implements prompt.Prompter for testing.
type mockPrompter struct {
	confirmed bool
	err       error
}

func (m *mockPrompter) Confirm(message, expectedValue string) (bool, error) {
	return m.confirmed, m.err
}

func TestConfirmThenRun_Force_SkipsPrompt(t *testing.T) {
	called := false
	err := confirmThenRun(nil, "warning", "value", true, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected action to be called when force is true")
	}
}

func TestConfirmThenRun_Confirmed_RunsAction(t *testing.T) {
	called := false
	p := &mockPrompter{confirmed: true}
	err := confirmThenRun(p, "warning", "value", false, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected action to be called when confirmed")
	}
}

func TestConfirmThenRun_Declined_ReturnsAborted(t *testing.T) {
	p := &mockPrompter{confirmed: false}
	err := confirmThenRun(p, "warning", "value", false, func() error {
		t.Fatal("action should not be called when declined")
		return nil
	})
	if !errors.Is(err, ErrOperationAborted) {
		t.Errorf("expected ErrOperationAborted, got: %v", err)
	}
}

func TestConfirmThenRun_PrompterError_Propagates(t *testing.T) {
	p := &mockPrompter{err: errors.New("input broken")}
	err := confirmThenRun(p, "warning", "value", false, func() error {
		t.Fatal("action should not be called on prompter error")
		return nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "reading confirmation input") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "input broken") {
		t.Errorf("expected original error in chain, got: %v", err)
	}
}

func TestConfirmThenRun_ActionError_Propagates(t *testing.T) {
	p := &mockPrompter{confirmed: true}
	actionErr := errors.New("delete failed")
	err := confirmThenRun(p, "warning", "value", false, func() error {
		return actionErr
	})
	if !errors.Is(err, actionErr) {
		t.Errorf("expected action error, got: %v", err)
	}
}

// Verify that confirmThenRun works with a real StandardPrompter backed by
// a bytes.Buffer, simulating actual user input.
func TestConfirmThenRun_WithStandardPrompter(t *testing.T) {
	t.Run("matching input confirms", func(t *testing.T) {
		input := strings.NewReader("my-bucket\n")
		var output strings.Builder
		p := prompt.NewStandardPrompter(input, &output)

		called := false
		err := confirmThenRun(p, "Delete my-bucket?", "my-bucket", false, func() error {
			called = true
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected action to be called")
		}
	})

	t.Run("mismatched input aborts", func(t *testing.T) {
		input := strings.NewReader("wrong-name\n")
		var output strings.Builder
		p := prompt.NewStandardPrompter(input, &output)

		err := confirmThenRun(p, "Delete my-bucket?", "my-bucket", false, func() error {
			t.Fatal("action should not be called on mismatch")
			return nil
		})
		if !errors.Is(err, ErrOperationAborted) {
			t.Errorf("expected ErrOperationAborted, got: %v", err)
		}
	})
}
