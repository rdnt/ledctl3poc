package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Table struct {
	cols []Column
	rows []Row
}

type Column struct {
	Title string
	Width int
}

type Row []string

func (m Table) String() string {
	renderedRows := make([]string, 0, len(m.rows))

	for i := range m.rows {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	return m.headersView() + "\n" + lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
}

func (m Table) headersView() string {
	var s = make([]string, 0, len(m.cols))
	for _, col := range m.cols {
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col.Title, col.Width, "…"))
		s = append(s, lipgloss.NewStyle().Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, s...)
}

func (m *Table) renderRow(rowID int) string {
	var s = make([]string, 0, len(m.cols))
	for i, value := range m.rows[rowID] {
		style := lipgloss.NewStyle().Width(m.cols[i].Width).MaxWidth(m.cols[i].Width).Inline(true)
		renderedCell := lipgloss.NewStyle().Render(style.Render(runewidth.Truncate(value, m.cols[i].Width, "…")))
		s = append(s, renderedCell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left, s...)

	return row
}
