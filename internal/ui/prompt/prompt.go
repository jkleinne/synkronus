// File: internal/ui/prompt/prompt.go
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Defines the interface for prompting the user for input
type Prompter interface {
	// Asks the user for confirmation by requiring them to type a specific expected value
	Confirm(message string, expectedValue string) (bool, error)
}

// Provides a standard implementation of the Prompter interface using specified input/output streams
type StandardPrompter struct {
	reader io.Reader
	writer io.Writer
}

// Creates a new StandardPrompter with the given input and output streams
func NewStandardPrompter(in io.Reader, out io.Writer) *StandardPrompter {
	return &StandardPrompter{
		reader: in,
		writer: out,
	}
}

// Asks the user for confirmation by requiring them to type a specific expected value
func (p *StandardPrompter) Confirm(message string, expectedValue string) (bool, error) {
	if expectedValue == "" {
		return false, fmt.Errorf("expected confirmation value cannot be empty")
	}

	fmt.Fprintln(p.writer, message)
	fmt.Fprintf(p.writer, "To confirm, please type the name '%s': ", expectedValue)

	reader := bufio.NewReader(p.reader)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return false, nil
		}
		return false, fmt.Errorf("error reading user input: %w", err)
	}

	cleanedInput := strings.TrimSpace(input)

	return cleanedInput == expectedValue, nil
}
