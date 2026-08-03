// Package cmd defines the park CLI command tree and wiring.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/render"
	"github.com/polymorcodeus/park/internal/store"
	"github.com/polymorcodeus/park/internal/theme"
	"github.com/urfave/cli/v3"
)

var (
	version   = "internal"
	buildTime string
)

// SetVersion sets the application version string used by the CLI.
func SetVersion(v string) {
	version = v
}

// SetBuildTime sets the build timestamp for version display.
func SetBuildTime(bt string) {
	buildTime = bt
}

func buildVersion() string {
	v := version
	if buildTime != "" {
		v += " (" + buildTime + ")"
	}
	return v
}

func Main() {
	var cfg *config.Config

	var (
		parkRoot, parkConfig string
		category             string
		newCategory          string
	)

	cmd := &cli.Command{
		Name:                  "park",
		Usage:                 "IPAA: a parking lot for markdown notes (Inbox/Projects/Areas/Archive)",
		Version:               buildVersion(),
		EnableShellCompletion: true,
		HideVersion:           false,
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		ErrWriter:             os.Stderr,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "park-root",
				Destination: &parkRoot,
				Value:       config.DefaultRootPath(),
				Sources:     cli.EnvVars("PARK_ROOT"),
				Usage:       "root directory for parked notes",
			},
			&cli.StringFlag{
				Name:        "park-config",
				Destination: &parkConfig,
				Value:       config.DefaultConfigPath(),
				Sources:     cli.EnvVars("PARK_CONFIG"),
				Usage:       "path to the config file",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			cfg = &config.Config{}
			if err := cfg.LoadConfig(parkRoot, parkConfig); err != nil {
				return ctx, styledExit(err, 1)
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			{
				Name:  "assist",
				Usage: "browse parked files and/or edit categories",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := assistPark(cfg); err != nil {
						return styledExit(err, 1)
					}
					return nil
				},
			},
			{
				Name:  "config",
				Usage: "print the default configuration file",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					out, err := config.DumpDefault(parkRoot)
					if err != nil {
						return styledExit(err, 1)
					}
					_, err = fmt.Fprint(cmd.Root().Writer, out)
					return err
				},
			},
			{
				Name:  "init",
				Usage: "create the category folders",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					msg, err := store.Init(cfg)
					if err != nil {
						return styledExit(err, 1)
					}
					if _, err := fmt.Fprintln(cmd.Root().Writer, msg); err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:      "new",
				Aliases:   []string{"add"},
				Usage:     "park a new note",
				ArgsUsage: "[filename]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "synopsis",
						Aliases: []string{"s"},
						Usage:   "one-line synopsis (what it is / why it matters)",
					},
					&cli.StringFlag{
						Name:    "source",
						Aliases: []string{"src"},
						Usage:   "where this came from (repo name, chat, etc.)",
					},
					&cli.StringFlag{
						Name:  "filename",
						Usage: "note filename (used for the file slug)",
					},
					&cli.StringFlag{
						Name:        "category",
						Aliases:     []string{"c"},
						Destination: &newCategory,
						Usage:       "category to park directly into if not the default",
					},
					&cli.StringFlag{
						Name:    "from-file",
						Aliases: []string{"f"},
						Usage:   "existing markdown file to move into the park",
					},
				},
				Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
					if newCategory != "" {
						if !cfg.HasCategory(newCategory) {
							err := fmt.Errorf("--category must be one of %s (got %q)", strings.Join(cfg.CategoryNames(), ", "), newCategory)
							return ctx, styledExit(err, 1)
						}
					}
					return ctx, nil
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := addPark(cfg, cmd); err != nil {
						return styledExit(err, 1)
					}
					return nil
				},
			},
			{
				Name:      "reclassify",
				Aliases:   []string{"recat"},
				Usage:     "reclassify a note",
				ArgsUsage: "<file>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "category",
						Aliases:     []string{"c"},
						Required:    true,
						Destination: &category,
						Usage:       "category to move the note into",
					},
				},
				Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
					if cmd.NArg() < 1 {
						return ctx, styledExit(fmt.Errorf("usage: park reclassify <file> --category <category>"), 2)
					}
					if !cfg.HasCategory(category) {
						return ctx, styledExit(fmt.Errorf("unknown category %q — valid: %s", category, strings.Join(cfg.CategoryNames(), ", ")), 2)
					}
					return ctx, nil
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := store.Reclassify(cfg, cmd.Args().First(), category); err != nil {
						return styledExit(err, 1)
					}
					return nil
				},
			},
			{
				Name:      "show",
				Usage:     "render a note to the terminal",
				ArgsUsage: "<file>",
				Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
					if cmd.NArg() < 1 {
						return ctx, styledExit(fmt.Errorf("usage: park show <file>"), 2)
					}
					return ctx, nil
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := render.ShowFile(store.ResolvePath(cfg, cmd.Args().First()), cmd.Root().Writer); err != nil {
						return styledExit(err, 1)
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, styledError(err))
		os.Exit(1)
	}
}

// styledExit wraps an error in the configured styled output and returns a
// CLI exit error with the supplied code.
func styledExit(err error, code int) error {
	return cli.Exit(styledError(err), code)
}

const errorBullet string = "󰯷" // "nf-md-alpha_e_box_outline

// StyledError returns a user-facing error string, styled when stdout is a
// terminal.
func styledError(e error) string {
	if e == nil {
		return ""
	}
	if !isTerminal(os.Stderr) {
		return e.Error()
	}
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmYellow)).
		Bold(true).
		Render("HEAVENS TO MURGATROYD!")
	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmRed)).
		Render(errorBullet, e.Error())
	return header + "\n" + body
}
