package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/model"
	"github.com/polymorcodeus/park/internal/note"
	"github.com/polymorcodeus/park/internal/render"
)

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// stdinIsTTY reports whether os.Stdin is connected to an interactive terminal.
func stdinIsTTY() bool {
	return isTerminal(os.Stdin)
}

// addPark parks a new note. It reads CLI input and delegates the creation
// decision to internal/note. If the input is incomplete, it opens the
// bubbletea form for the missing metadata.
func addPark(cfg *config.Config, cmd *cli.Command) error {
	in := note.NoteInput{
		Filename: cmd.String("filename"),
		Synopsis: cmd.String("synopsis"),
		Source:   cmd.String("source"),
		Category: cmd.String("category"),
		FromFile: cmd.String("from-file"),
	}
	if in.Filename == "" {
		in.Filename = cmd.Args().First()
	}

	if in.FromFile != "" {
		data, err := os.ReadFile(in.FromFile)
		if err != nil {
			return fmt.Errorf("from-file: %w", err)
		}
		in.Body = string(data)
	} else if !stdinIsTTY() {
		data, err := io.ReadAll(cmd.Reader)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		in.Body = string(data)
	}

	outcome, err := note.AddNote(cfg, in)
	if err != nil {
		return err
	}
	if outcome.Path != "" {
		if _, err := fmt.Fprintln(cmd.Root().Writer, "parked:", outcome.Path); err != nil {
			return err
		}
		return nil
	}
	return runNoteForm(cfg, cmd, outcome.Form)
}

func runNoteForm(cfg *config.Config, cmd *cli.Command, form *note.NoteForm) error {
	m, err := model.NewNoteFormModel(cfg, form.Filename, form.Synopsis, form.Source, form.Category, form.Body, form.FromFile)
	if err != nil {
		return err
	}

	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return err
	}

	final, ok := result.(model.NoteFormModel)
	if !ok {
		return fmt.Errorf("unexpected model type from form")
	}
	res := final.Result()
	if res.Err != nil {
		return res.Err
	}
	if res.Path != "" {
		if _, err := fmt.Fprintln(cmd.Root().Writer, "parked:", res.Path); err != nil {
			return err
		}
	}
	return nil
}

// formatInitMessage formats the result of store.Init for user-facing output.
func formatInitMessage(created, existed []string) string {
	if len(created) == 0 {
		return "all park folders already exist"
	}

	msg := fmt.Sprintf("created park folders: %s", strings.Join(created, ", "))
	if len(existed) > 0 {
		msg += fmt.Sprintf(" (%s already existed)", strings.Join(existed, ", "))
	}
	return msg
}

func assistPark(cfg *config.Config) error {
	m, err := model.NewAssistModel(cfg)
	if err != nil {
		return err
	}

	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return err
	}

	final, ok := result.(model.AssistModel)
	if !ok {
		return fmt.Errorf("unexpected model type from assist")
	}
	if final.ViewFile != "" {
		if err := render.ShowFile(final.ViewFile, os.Stdout); err != nil {
			return err
		}
	}
	return nil
}
