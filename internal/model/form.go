// Package model provides bubbletea TUI models for park.
package model

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/fs"
	"github.com/polymorcodeus/park/internal/note"
)

const maxFilePreviewLines = 6

// noteKeyMap defines the key bindings for the note form.
type noteKeyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Enter    key.Binding
	Submit   key.Binding
	Cancel   key.Binding
	CatLeft  key.Binding
	CatRight key.Binding
}

var noteKeys = noteKeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous field"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "next/submit"),
	),
	Submit: key.NewBinding(
		key.WithKeys("ctrl+enter"),
		key.WithHelp("ctrl+enter", "submit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc/ctrl+c", "cancel"),
	),
	CatLeft: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "previous category"),
	),
	CatRight: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next category"),
	),
}

// NoteFormModel is a bubbletea form for interactively creating or ingesting a
// parked note. Pre-filled values come from CLI flags; missing fields are
// collected here.
type NoteFormModel struct {
	cfg             *config.Config
	inputs          []textinput.Model
	bodyInput       textarea.Model
	categoryIdx     int
	focusIndex      int
	fromFile        string
	filePreview     string
	bodyCursorReset bool
	err             error
	createdPath     string
	styles          *styles
	keys            noteKeyMap
	width           int
}

// NoteFormResult is returned after the form exits.
type NoteFormResult struct {
	Path string
	Err  error
}

// textInputStyles returns the themed textinput styles used by the form.
func (s *styles) textInputStyles() textinput.Styles {
	st := textinput.New().Styles()
	st.Focused.Prompt = s.focusedPrompt
	st.Focused.Text = s.focusedText
	st.Blurred.Prompt = s.blurredPrompt
	st.Blurred.Text = s.blurredText
	return st
}

// NewNoteFormModel builds a form model with the supplied initial values.
func NewNoteFormModel(cfg *config.Config, filename, synopsis, source, category, body, fromFile string) (NoteFormModel, error) {
	idx := -1
	for i, cl := range cfg.Categories {
		if cl.Name == category {
			idx = i
			break
		}
	}
	if idx < 0 {
		return NoteFormModel{}, fmt.Errorf("unknown category %q", category)
	}

	s := newStyles()

	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		ti := textinput.New()
		ti.CharLimit = 240
		ti.SetWidth(40)
		ti.SetStyles(s.textInputStyles())

		switch i {
		case 0:
			ti.Prompt = "filename: "
			ti.Placeholder = "note filename"
			ti.SetValue(filename)
		case 1:
			ti.Prompt = "synopsis: "
			ti.Placeholder = "one-line description"
			ti.SetValue(synopsis)
		case 2:
			ti.Prompt = "source:   "
			ti.Placeholder = "where this came from"
			ti.SetValue(source)
		}
		inputs[i] = ti
	}

	ta := textarea.New()
	ta.Placeholder = "note body (optional)"
	ta.SetValue(body)
	ta.SetStyles(textarea.DefaultStyles(true))
	ta.MaxHeight = 10

	var preview string
	if fromFile != "" {
		preview = previewFile(fromFile)
	}

	m := NoteFormModel{
		cfg:             cfg,
		inputs:          inputs,
		bodyInput:       ta,
		categoryIdx:     idx,
		focusIndex:      0,
		fromFile:        fromFile,
		filePreview:     preview,
		bodyCursorReset: body != "",
		styles:          s,
		keys:            noteKeys,
		width:           minWidth,
	}
	m, _ = m.updateFocus()
	return m, nil
}

// hasBodyField reports whether the form includes an editable body textarea.
func (m NoteFormModel) hasBodyField() bool {
	return m.fromFile == ""
}

func (m NoteFormModel) bodyIndex() int {
	if !m.hasBodyField() {
		return -1
	}
	return 3
}

func (m NoteFormModel) categoryIndex() int {
	if !m.hasBodyField() {
		return 3
	}
	return 4
}

func (m NoteFormModel) submitIndex() int {
	if !m.hasBodyField() {
		return 4
	}
	return 5
}

func (m NoteFormModel) lastIndex() int {
	return m.submitIndex()
}

// updateFocus applies focus/blur to every field based on the current focus
// index and returns the collected commands.
func (m NoteFormModel) updateFocus() (NoteFormModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 6)
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds = append(cmds, m.inputs[i].Focus())
			continue
		}
		m.inputs[i].Blur()
	}
	if m.hasBodyField() && m.focusIndex == m.bodyIndex() {
		cmds = append(cmds, m.bodyInput.Focus())
		if m.bodyCursorReset {
			m.bodyInput.CursorStart()
			m.bodyCursorReset = false
		}
	} else {
		m.bodyInput.Blur()
	}
	return m, tea.Batch(cmds...)
}

// updateInputs forwards the message to every field. Only focused fields
// respond to character input, so it is safe to update all of them.
func (m NoteFormModel) updateInputs(msg tea.Msg) (NoteFormModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 6)
	for i := range m.inputs {
		updated, cmd := m.inputs[i].Update(msg)
		m.inputs[i] = updated
		cmds = append(cmds, cmd)
	}
	updatedBody, cmd := m.bodyInput.Update(msg)
	m.bodyInput = updatedBody
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m NoteFormModel) filename() string { return strings.TrimSpace(m.inputs[0].Value()) }
func (m NoteFormModel) synopsis() string { return strings.TrimSpace(m.inputs[1].Value()) }
func (m NoteFormModel) source() string   { return strings.TrimSpace(m.inputs[2].Value()) }
func (m NoteFormModel) categoryName() string {
	return m.cfg.Categories[m.categoryIdx].Name
}

func (m NoteFormModel) canSubmit() bool {
	return m.filename() != "" && m.synopsis() != "" && m.source() != ""
}

func (m NoteFormModel) body() string {
	if m.fromFile != "" {
		return ""
	}
	return strings.TrimRight(m.bodyInput.Value(), "\n")
}

// previewFile returns the first few lines of a file for display in the form.
func previewFile(path string) string {
	path = fs.ExpandPath(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("(unable to read %s: %v)", path, err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > maxFilePreviewLines {
		lines = lines[:maxFilePreviewLines]
		lines = append(lines, "...")
	}
	return strings.Join(lines, "\n")
}

// submitMsg signals that the form should be submitted.
type submitMsg struct{}

func (m NoteFormModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		return submitMsg{}
	}
}

func (m NoteFormModel) createCmd() tea.Cmd {
	filename := m.filename()
	synopsis := m.synopsis()
	source := m.source()
	category := m.categoryName()
	body := m.body()
	fromFile := m.fromFile
	cfg := m.cfg

	return func() tea.Msg {
		var path string
		var err error
		if fromFile != "" {
			path, err = note.IngestFile(cfg, fromFile, filename, synopsis, source, category, body)
		} else {
			path, err = note.NewWithBody(cfg, filename, synopsis, source, category, body)
		}
		if err != nil {
			return formErrorMsg{err: err}
		}
		return formCreatedMsg{path: path}
	}
}

type formCreatedMsg struct{ path string }
type formErrorMsg struct{ err error }

// Init implements tea.Model.
func (m NoteFormModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink, tea.RequestBackgroundColor)
}

// Update implements tea.Model.
func (m NoteFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(minWidth, msg.Width)
		for i := range m.inputs {
			m.inputs[i].SetWidth(msg.Width - 20)
		}
		m.bodyInput.SetWidth(msg.Width - 20)
		m.bodyInput.SetHeight(minHeight)

	case tea.BackgroundColorMsg:
		m.bodyInput.SetStyles(textarea.DefaultStyles(msg.IsDark()))

	case formCreatedMsg:
		m.createdPath = msg.path
		return m, tea.Quit

	case formErrorMsg:
		m.err = msg.err
		return m, nil

	case submitMsg:
		if !m.canSubmit() {
			m.err = fmt.Errorf("filename, synopsis, and source are required")
			return m, nil
		}
		return m, m.createCmd()

	case tea.KeyPressMsg:
		m.err = nil
		switch {
		case key.Matches(msg, m.keys.Cancel):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Tab):
			m.focusIndex++
			if m.focusIndex > m.lastIndex() {
				m.focusIndex = 0
			}
			if !m.hasBodyField() && m.focusIndex == m.bodyIndex() {
				m.focusIndex++
			}
			m, cmd := m.updateFocus()
			return m, cmd

		case key.Matches(msg, m.keys.ShiftTab):
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = m.lastIndex()
			}
			if !m.hasBodyField() && m.focusIndex == m.bodyIndex() {
				m.focusIndex--
			}
			m, cmd := m.updateFocus()
			return m, cmd

		case key.Matches(msg, m.keys.Submit):
			return m, m.submitCmd()

		case key.Matches(msg, m.keys.Enter):
			if !m.hasBodyField() || m.focusIndex != m.bodyIndex() {
				if m.focusIndex == m.submitIndex() || m.focusIndex == m.categoryIndex() {
					return m, m.submitCmd()
				}
				m.focusIndex++
				if m.focusIndex > m.lastIndex() {
					m.focusIndex = 0
				}
				if !m.hasBodyField() && m.focusIndex == m.bodyIndex() {
					m.focusIndex++
				}
				m, cmd := m.updateFocus()
				return m, cmd
			}

		case key.Matches(msg, m.keys.CatLeft):
			if m.focusIndex == m.categoryIndex() {
				m.categoryIdx--
				if m.categoryIdx < 0 {
					m.categoryIdx = len(m.cfg.Categories) - 1
				}
				return m, nil
			}

		case key.Matches(msg, m.keys.CatRight):
			if m.focusIndex == m.categoryIndex() {
				m.categoryIdx++
				if m.categoryIdx >= len(m.cfg.Categories) {
					m.categoryIdx = 0
				}
				return m, nil
			}
		}
	}

	m, cmd := m.updateInputs(msg)
	return m, cmd
}

// View implements tea.Model.
func (m NoteFormModel) View() tea.View {
	if m.styles == nil {
		return tea.NewView("")
	}

	s := m.styles
	var b strings.Builder

	// Category tabs. The selected category is always highlighted so users
	// know which category the note will be parked in. When the category area
	// itself is focused, the selected tab is rendered with a visible indicator.
	var renderedTabs []string
	for i, cl := range m.cfg.Categories {
		style := s.inactiveTab
		label := cl.Name
		if i == m.categoryIdx {
			style = s.activeTab
			if m.focusIndex == m.categoryIndex() {
				label = "❯ " + cl.Name + " ❮"
			}
		}
		renderedTabs = append(renderedTabs, style.Render(label))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	gapWidth := max(0, m.width-lipgloss.Width(row))
	gap := s.topBorder.Render(strings.Repeat(" ", gapWidth))
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Text inputs.
	for _, in := range m.inputs {
		b.WriteString(in.View())
		b.WriteString("\n")
	}

	// Body textarea or file preview.
	if m.hasBodyField() {
		b.WriteString(s.highlight.Render("body:"))
		b.WriteString("\n")
		b.WriteString(m.bodyInput.View())
		b.WriteString("\n")
	} else {
		b.WriteString(s.highlight.Render("from-file: "))
		b.WriteString(m.fromFile)
		b.WriteString("\n")
		b.WriteString(m.filePreview)
		b.WriteString("\n")
	}

	// Submit button.
	button := m.renderSubmitButton()
	b.WriteString("\n")
	b.WriteString(s.window.Width(m.width).Render(button))
	b.WriteString("\n")

	// Help / error.
	help := "tab/shift+tab: move  ·  ←/→: change category  ·  enter: next/submit  ·  esc: cancel"
	// b.WriteString(s.helpText.Render(help))

	b.WriteString(s.window.Width(m.width).Render(s.helpText.Render(help)))

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(s.errorText.Render(m.err.Error()))
	}

	v := tea.NewView(s.doc.Render(b.String()))
	v.AltScreen = true
	return v
}

func (m NoteFormModel) renderSubmitButton() string {
	label := "󰄽 Submit 󰄾"
	if m.focusIndex == m.submitIndex() {
		return m.styles.submitButtonFocus.Render(label)
	}
	return m.styles.submitButton.Render(label)
}

// Result returns the outcome of the form after the program exits.
func (m NoteFormModel) Result() NoteFormResult {
	return NoteFormResult{Path: m.createdPath, Err: m.err}
}
