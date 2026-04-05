package main

import (
	"fmt"
	"synkronus/internal/ui/prompt"
)

// confirmThenRun prompts the user for confirmation unless force is true,
// then runs the action. Returns ErrOperationAborted if the user declines.
func confirmThenRun(prompter prompt.Prompter, message, expectedValue string, force bool, action func() error) error {
	if !force {
		confirmed, err := prompter.Confirm(message, expectedValue)
		if err != nil {
			return fmt.Errorf("reading confirmation input: %w", err)
		}
		if !confirmed {
			fmt.Println("Deletion aborted: Confirmation mismatch or cancelled.")
			return ErrOperationAborted
		}
	}
	return action()
}
