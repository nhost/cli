package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

type Application struct {
	stdout io.Writer
	stderr io.Writer
}

func NewApplication(cCtx *cli.Context) *Application {
	return &Application{
		stdout: cCtx.App.Writer,
		stderr: cCtx.App.ErrWriter,
	}
}

func (app *Application) Println(msg string, a ...any) {
	if _, err := fmt.Fprintln(app.stdout, fmt.Sprintf(msg, a...)); err != nil {
		panic(err)
	}
}

func (app *Application) Infoln(msg string, a ...any) {
	if _, err := fmt.Fprintln(app.stdout, info(fmt.Sprintf(msg, a...))); err != nil {
		panic(err)
	}
}

func (app *Application) Warnln(msg string, a ...any) {
	if _, err := fmt.Fprintln(app.stdout, warn(fmt.Sprintf(msg, a...))); err != nil {
		panic(err)
	}
}

func (app *Application) PromptMessage(msg string, a ...any) {
	if _, err := fmt.Fprint(app.stdout, promptMessage("- "+fmt.Sprintf(msg, a...))); err != nil {
		panic(err)
	}
}

func (app *Application) PromptInput(hide bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var response string
	var err error

	if !hide {
		response, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
	} else {
		output, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		response = string(output)
	}

	return strings.TrimSpace(response), err
}
