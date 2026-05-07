package main

import (
	"fmt"
	"os"

	"github.com/0xtbug/evm-disperse-tools/internal/bootstrap"
	tea "github.com/charmbracelet/bubbletea"
)

func run() error {
	app, err := bootstrap.BuildApp()
	if err != nil {
		return fmt.Errorf("failed to bootstrap app: %w", err)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
