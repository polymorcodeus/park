package cmd

import (
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/model"
	"github.com/polymorcodeus/park/internal/note"
	"github.com/polymorcodeus/park/internal/render"
)

// isTerminal reports whether the given file descriptor is connected to an
// interactive terminal. It is used for both stdin (see stdinIsTTY) and stderr
// (for styled error output).
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

// draftFromCmd builds a note.Draft from the CLI flags and positional args.
func draftFromCmd(cmd *cli.Command) note.Draft {
	d := note.Draft{
		Filename: cmd.String("filename"),
		Metadata: note.Metadata{
			Synopsis: cmd.String("synopsis"),
			Source:   cmd.String("source"),
			Category: cmd.String("category"),
		},
		FromFile: cmd.String("from-file"),
	}
	if d.Filename == "" {
		d.Filename = cmd.Args().First()
	}
	return d
}

// printParked writes the standard "parked: <path>" confirmation.
func printParked(w io.Writer, path string) error {
	if _, err := fmt.Fprintln(w, "parked:", path); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

// addPark parks a new note. It reads CLI input and delegates the creation
// decision to internal/note. If the input is incomplete, it opens the
// bubbletea form for the missing metadata.
func addPark(cfg *config.Config, cmd *cli.Command, w io.Writer) error {
	d := draftFromCmd(cmd)

	if !stdinIsTTY() {
		data, err := io.ReadAll(cmd.Reader)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		d.Body = string(data)
	}

	outcome, err := note.Add(cfg, d)
	if err != nil {
		return err
	}
	if outcome.Path != "" {
		return printParked(w, outcome.Path)
	}
	return runNoteForm(cfg, w, outcome.Form)
}

func runNoteForm(cfg *config.Config, w io.Writer, seed *note.Draft) error {
	m, err := model.NewNoteFormModel(cfg, *seed)
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
		return printParked(w, res.Path)
	}
	return nil
}

func assistPark(cfg *config.Config, w io.Writer) error {
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
		if err := render.ShowFile(final.ViewFile, w); err != nil {
			return err
		}
	}
	return nil
}
