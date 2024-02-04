package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ledctl3/pkg/uuid"
)

func runTUI() {
	columns := []table.Column{
		{Title: "Input", Width: 8},
		{Title: "Source", Width: 8},
		{Title: "Sink", Width: 8},
		{Title: "Output", Width: 8},
	}

	rows := []table.Row{
		{"(empty)"},
		//{"f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184"},
		//{"f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184"},
		//{"f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184"},
		//{"f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184", "f5b8bc16-bc2a-4a7d-9fde-b55fc4217184"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(1),
	)

	s := table.DefaultStyles()

	//s.Header = s.Header.Border(lipgloss.HiddenBorder())
	//s.Header = s.Header.
	//	BorderStyle(lipgloss.HiddenBorder())
	//s.Cell = s.Cell.BorderStyle(lipgloss.Border{
	//	Top:          "─",
	//	Bottom:       "─",
	//	Left:         "│",
	//	Right:        "│",
	//	TopLeft:      "┌",
	//	TopRight:     "┐",
	//	BottomLeft:   "└",
	//	BottomRight:  "┘",
	//	MiddleLeft:   ">",
	//	MiddleRight:  ">",
	//	Middle:       "─",
	//	MiddleTop:    "@",
	//	MiddleBottom: "─",
	//})
	//	BorderForeground(lipgloss.Color("240")).
	//	BorderBottom(true).
	//	Bold(false)
	//s.Cell.Border()
	//s.Cell.Border(lipgloss.HiddenBorder(), false, false, false, false)
	s.Header.Foreground(lipgloss.Color("#cdd6f4")).PaddingBottom(1)
	s.Header.Padding(0, 2, 1)
	//s.Header.Padding(0, 0, 1)
	//s.Cell.Foreground(lipgloss.Color("#6c7086"))
	s.Cell.Padding(0, 2, 0)
	s.Selected.Foreground(lipgloss.Color("#a6e3a1"))
	//s.Selected = s.Selected.
	//	Foreground(lipgloss.Color("229")).
	//	Background(lipgloss.Color("57")).
	//	Bold(false)
	//	t.SetStyles(table.Styles{
	//	Header:   lipgloss.NewStyle(),
	//	Cell:     lipgloss.NewStyle().Border(lipgloss.HiddenBorder()),
	//	Selected: lipgloss.NewStyle().Border(lipgloss.HiddenBorder()),
	//})
	t.SetStyles(s)

	h := help.New()
	h.ShowAll = true

	h.Styles.FullKey.Foreground(lipgloss.Color("#cdd6f4"))
	h.Styles.FullDesc.Foreground(lipgloss.Color("#6c7086"))

	//h.ShortSeparator = "\t\t"
	//h.FullSeparator = "\t\t"
	//h.ShowAll = true

	km := keys

	m := model{
		events:      make(chan any),
		table:       t,
		help:        h,
		keymap:      km,
		tableStyles: s,
		io: &io{
			ids:  make(map[string]bool),
			cols: columns,
		},
	}

	p := tea.NewProgram(
		m,
		//tea.WithAltScreen(),
		//tea.WithMouseAllMotion(),
		tea.WithMouseCellMotion(),
	)

	go func() {
		for i := 0; i < 2; i++ {
			time.Sleep(500 * time.Millisecond)
			m.events <- IoEntryAdded{
				InputId:    uuid.New().String(),
				InputName:  "potato",
				SourceId:   uuid.New().String(),
				SourceName: "",
				SinkId:     uuid.New().String(),
				SinkName:   "some random sink name",
				OutputId:   uuid.New().String(),
				OutputName: "",
			}

			time.Sleep(500 * time.Millisecond)
			m.events <- IoEntryAdded{
				InputId:    uuid.New().String(),
				InputName:  "tomato",
				SourceId:   uuid.New().String(),
				SourceName: "",
				SinkId:     uuid.New().String(),
				SinkName:   "",
				OutputId:   uuid.New().String(),
				OutputName: "",
			}

			time.Sleep(500 * time.Millisecond)
			m.events <- IoEntryAdded{
				InputId:    uuid.Nil.String(),
				InputName:  "banana banananananana",
				SourceId:   uuid.Nil.String(),
				SourceName: "",
				SinkId:     uuid.Nil.String(),
				SinkName:   "some random sink name",
				OutputId:   uuid.Nil.String(),
				OutputName: "",
			}

			time.Sleep(500 * time.Millisecond)
			m.events <- IoEntryAdded{
				InputId:    uuid.Nil.String(),
				InputName:  "banana",
				SourceId:   uuid.Nil.String(),
				SourceName: "nananananana",
				SinkId:     uuid.Nil.String(),
				SinkName:   "",
				OutputId:   uuid.Nil.String(),
				OutputName: "smol",
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	Padding(1, 1, 1).
	Foreground(lipgloss.Color("#a6adc8")).
	BorderForeground(lipgloss.Color("#cdd6f4"))

type model struct {
	events      chan any
	state       string
	table       table.Model
	tableStyles table.Styles
	help        help.Model
	keymap      keymap
	width       int
	height      int
	renders     int
	io          *io
	debug       []string
}

type keymap struct {
	quit key.Binding
}

var keys = keymap{
	quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("[^c]", "quit"),
	),
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.quit},
	}
}

//func (k keymap) ShortHelp() []key.Binding {
//	return []key.Binding{
//		key.NewBinding(key.WithKeys("esc, q"), key.WithHelp("esc", "quit")),
//	}
//}
//
//func (k keymap) FullHelp() [][]key.Binding {
//	return [][]key.Binding{k.ShortHelp()}
//}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScrollArea,
		//tick(),
		nextEvent(m.events),
		//randomRows,
		//ticker(randomRows),
	)
}

//func ticker(f func() tea.Msg) tea.Cmd {
//	return tea.Tick(
//		1000*time.Millisecond, func(t time.Time) tea.Msg {
//			return f()
//		},
//	)
//}

type IoEvent struct {
}

type IoEntryAdded struct {
	InputId    string
	InputName  string
	SourceId   string
	SourceName string
	SinkId     string
	SinkName   string
	OutputId   string
	OutputName string
}

func nextEvent(sub chan any) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func tick() tea.Cmd {
	return tea.Tick(
		1000*time.Millisecond, func(t time.Time) tea.Msg {
			return t
		},
	)
}

type io struct {
	rows []ioRow
	ids  map[string]bool
	cols []table.Column
}

type ioRow struct {
	id     string
	input  string
	source string
	sink   string
	output string
}

func (io *io) tableRows(selected int) []table.Row {
	if len(io.rows) == 0 {
		return []table.Row{{"(empty)"}}
	}

	rows := make([]table.Row, len(io.rows))
	for i, r := range io.rows {
		rows[i] = []string{r.input, r.source, r.sink, r.output}

		//for j := 0; j < 4; j++ {
		//	if i == selected && j == 0 {
		//		rows[i][j] = "> " + rows[i][j]
		//	} else {
		//		rows[i][j] = "  " + rows[i][j]
		//	}
		//	if j == 3 {
		//		rows[i][j] = rows[i][j] + "  "
		//	}
		//}

	}
	return rows
}

func (io *io) handleIoEntryAdded(e IoEntryAdded) {
	id := fmt.Sprintf("%s:%s:%s:%s", e.InputId, e.SourceId, e.SinkId, e.OutputId)
	if _, ok := io.ids[id]; ok {
		return
	}

	io.ids[id] = true

	r := ioRow{
		id: id,
	}

	if e.InputName != "" {
		r.input = e.InputName
	} else {
		r.input = fmt.Sprintf("{%s}", e.InputId)
	}

	if e.SourceName != "" {
		r.source = e.SourceName
	} else {
		r.source = fmt.Sprintf("{%s}", e.SourceId)
	}

	if e.SinkName != "" {
		r.sink = e.SinkName
	} else {
		r.sink = fmt.Sprintf("{%s}", e.SinkId)
	}

	if e.OutputName != "" {
		r.output = e.OutputName
	} else {
		r.output = fmt.Sprintf("{%s}", e.OutputId)
	}

	io.rows = append(io.rows, r)
	io.cols[0].Width = min(max(io.cols[0].Width, len(r.input)), 38)
	io.cols[1].Width = min(max(io.cols[1].Width, len(r.source)), 38)
	io.cols[2].Width = min(max(io.cols[2].Width, len(r.sink)), 38)
	io.cols[3].Width = min(max(io.cols[3].Width, len(r.output)), 38)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.debug = nil

	//var render bool
	//invalidate := func() {
	//	render = true
	//}

	switch msg := msg.(type) {

	case IoEvent:
		//cmds = append(cmds, randomRows)
		cmds = append(cmds, nextEvent(m.events))

	case IoEntryAdded:
		m.io.handleIoEntryAdded(msg)

		m.table.SetRows(m.io.tableRows(m.table.Cursor()))
		m.table.SetColumns(m.io.cols)

		m.updateTable()

		//m.table.SetHeight(max(min(len(m.io.rows), m.height-6), 1))
		//invalidate()

		cmds = append(cmds, nextEvent(m.events))

	case tea.KeyMsg:

		switch {
		case key.Matches(msg, m.table.KeyMap.LineUp):
			cur := max(m.table.Cursor()-1, 0)
			m.debug = append(m.debug, "cursor", fmt.Sprint(cur))
			m.table.SetRows(m.io.tableRows(cur))
			m.updateTable()

		case key.Matches(msg, m.table.KeyMap.LineDown):
			cur := min(m.table.Cursor()+1, len(m.io.rows)-1)
			m.debug = append(m.debug, "cursor", fmt.Sprint(cur))
			m.table.SetRows(m.io.tableRows(cur))
			m.updateTable()
		}

		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}

	case rowsEvent:
		m.table.SetRows(msg)

		m.updateTable()

		//m.table.SetHeight(max(min(len(m.io.rows), m.height-6), 1))
		//invalidate()

	case tea.MouseMsg:
		//fmt.Println("MouseMsg", msg)

	case time.Time:
		//cmds = append(cmds, tick())
		//cmds = append(cmds, randomRows)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.updateTable()
		//m.table.SetWidth(msg.Width)
		//m.table.SetHeight(max(min(len(m.io.rows), m.height), 1))
		//invalidate()
		//fmt.Println("WindowSizeMsg", msg)

	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)
	cmds = append(cmds, tableCmd)

	//if render {
	//	m.renders++
	//	cmds = append(cmds, func() tea.Msg {
	//		return tea.Tick(0, func(t time.Time) tea.Msg {
	//			return t
	//		})
	//	})
	//}

	return m, tea.Batch(cmds...)
}

func (m *model) updateTable() {
	if len(m.io.rows) == 0 {
		//m.tableStyles.Cell = m.tableStyles.Cell.Foreground(lipgloss.Color("#6c7086"))
		m.tableStyles.Selected = m.tableStyles.Selected.Foreground(lipgloss.Color("#6c7086")).UnsetBackground()
		m.table.SetStyles(m.tableStyles)
	} else {
		m.tableStyles.Selected = m.tableStyles.Selected.
			//Foreground(lipgloss.Color("#a6e3a1")).
			//UnsetBackground().
			Background(lipgloss.Color("#2a2b3c")).
			Foreground(lipgloss.Color("#cdd6f4"))
		m.table.SetStyles(m.tableStyles)
	}

	m.table.SetWidth(m.width)
	m.table.SetHeight(max(min(len(m.io.rows), m.height), 1))
}

var renders int

func (m model) View() string {
	var v strings.Builder
	v.Grow(m.width * m.height)

	v.WriteString(baseStyle.Render(m.table.View()))
	v.WriteString("\n")
	v.WriteString(m.help.View(m.keymap))
	v.WriteString("\n")
	v.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render(fmt.Sprintf("debug: frame %d %s", renders, strings.Join(m.debug, " "))))
	renders++

	return v.String()
}

type rowsEvent []table.Row

func randomRows() tea.Msg {
	var rows []table.Row

	for i := 0; i < rand.Intn(9)+1; i++ {
		var r []string
		for j := 0; j < 4; j++ {
			r = append(r, uuid.New().String())
		}
		rows = append(rows, r)
	}

	return rowsEvent(rows)
}
