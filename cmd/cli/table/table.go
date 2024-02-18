package table

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Table struct {
	hasHeaders bool
	rows       [][]string
	width      []int
	padding    int
}

func New() *Table {
	return &Table{}
}

func (t *Table) WithHeaders(headers []string) *Table {
	if !t.hasHeaders {
		t.rows = append([][]string{headers}, t.rows...)
		t.hasHeaders = true
	} else {
		t.rows[0] = headers
	}

	t.recalculateWidth()

	return t
}

func (t *Table) WithRows(rows [][]string) *Table {
	if len(rows) == 0 {
		rows = [][]string{{"(empty)"}}
	}

	if t.hasHeaders {
		t.rows = append(t.rows[0:1], rows...)
	} else {
		t.rows = rows
	}

	t.recalculateWidth()

	return t
}

func (t *Table) WithPadding(padding int) *Table {
	t.padding = padding

	t.recalculateWidth()

	return t
}

func (t *Table) String() string {
	renderedRows := make([]string, len(t.rows))

	for i, row := range t.rows {
		renderedRows[i] = t.renderRow(row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
}

func (t *Table) renderRow(cols []string) string {
	var s = make([]string, 0, len(cols))
	for i, value := range cols {
		style := lipgloss.NewStyle().Width(t.width[i]).MaxWidth(t.width[i]).Inline(true)
		renderedCell := lipgloss.NewStyle().Render(style.Render(runewidth.Truncate(value, t.width[i], "…")))
		s = append(s, renderedCell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left, s...)

	return row
}

func (t *Table) recalculateWidth() {
	t.width = nil

	for _, row := range t.rows {
		for j, cell := range row {
			if len(t.width) <= j {
				t.width = append(t.width, 0)
			}

			cellWidth := runewidth.StringWidth(cell) + t.padding
			if cellWidth > t.width[j] {
				t.width[j] = cellWidth
			}
		}
	}
}
