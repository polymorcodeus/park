package model

import (
	"charm.land/lipgloss/v2"
	"github.com/polymorcodeus/park/internal/theme"
)

const (
	minWidth  = 120
	minHeight = 20
)

type styles struct {
	doc               lipgloss.Style
	highlight         lipgloss.Style
	inactiveTab       lipgloss.Style
	activeTab         lipgloss.Style
	topBorder         lipgloss.Style
	window            lipgloss.Style
	errorText         lipgloss.Style
	helpText          lipgloss.Style
	submitButton      lipgloss.Style
	submitButtonFocus lipgloss.Style
	focusedPrompt     lipgloss.Style
	focusedText       lipgloss.Style
	blurredPrompt     lipgloss.Style
	blurredText       lipgloss.Style
	listNormalTitle   lipgloss.Style
	listNormalDesc    lipgloss.Style
	listSelectedTitle lipgloss.Style
	listSelectedDesc  lipgloss.Style
	listDimmedTitle   lipgloss.Style
	listDimmedDesc    lipgloss.Style
}

func newStyles() *styles {
	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")

	s := new(styles)
	s.doc = lipgloss.NewStyle().
		Padding(1, 2)
	s.inactiveTab = lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(lipgloss.Color(theme.CharmPink)).
		Foreground(lipgloss.Color(theme.CharmTextMute)).
		Padding(0, 1)
	s.activeTab = lipgloss.NewStyle().
		Border(activeTabBorder, true).
		BorderForeground(lipgloss.Color(theme.CharmPink)).
		Background(lipgloss.Color(theme.CharmPink)).
		Foreground(lipgloss.Color(theme.CharmBG)).
		Padding(0, 1)
	s.topBorder = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(theme.CharmPink))
	s.highlight = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmPink))
	s.window = lipgloss.NewStyle().
		BorderBottom(true).
		Padding(1, 2).
		Align(lipgloss.Center).
		UnsetBorderTop()
	s.errorText = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmRed))
	s.helpText = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextFaint))
	s.submitButton = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextFaint))
	s.submitButtonFocus = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmBG)).
		Background(lipgloss.Color(theme.CharmPink))
	s.focusedPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmPink))
	s.focusedText = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmText))
	s.blurredPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextFaint))
	s.blurredText = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmText))
	s.listNormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmText))
	s.listNormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextFaint))
	s.listSelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmPurpleLt))
	s.listSelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextMute))
	s.listDimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextFaint))
	s.listDimmedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.CharmTextFaint))
	return s
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	b := lipgloss.RoundedBorder()
	b.BottomLeft = left
	b.Bottom = middle
	b.BottomRight = right
	return b
}
