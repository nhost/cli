package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func UserInput(wr io.Writer, message string, hide bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var response string
	var err error

	if _, err := fmt.Fprintf(wr, message+": "); err != nil {
		return "", fmt.Errorf("failed to write to writer: %w", err)
	}
	if !hide {
		response, err = reader.ReadString('\n')
	} else {
		output, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		response = string(output)
	}

	return strings.TrimSpace(response), err
}
