package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"ledctl3/cmd/registry/cobrautil"
)

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Background(lipgloss.Color("#444"))
	ti.Focus()
	ti.SetValue("")
	ti.ShowSuggestions = true
	//ti.Cursor.Blink = false
	//ti.SetSuggestions(cmds)
	ti.Cursor.BlinkSpeed = 10 * time.Second
	//ti.SetCursor(1)

	return ti
}

func runTUIText(root *cobra.Command) {

	//ti := textinput.New()
	//ti.Placeholder = "<name>"
	//ti.Prompt = ""
	//ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))
	//ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Background(lipgloss.Color("#f00"))
	//ti.Focus()
	//ti.CharLimit = 50
	////ti.Width = 20
	//ti.ShowSuggestions = true

	m := modelText{
		root: root,
		//inputs: []textinput.Model{newTextInput()},
		input: newTextInput(),
		//ti
		//textInput: ti,
	}

	p := tea.NewProgram(
		m,
		//tea.WithAltScreen(),
		//tea.WithMouseAllMotion(),
		//tea.WithMouseCellMotion(),
	)

	go func() {

	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type modelText struct {
	root  *cobra.Command
	input textinput.Model

	//Cmd string
	//textInput textinput.Model
	//hint  string
	debug    string
	curr     string
	profName string
	lex      []string
}

func (m modelText) Init() tea.Cmd {
	return tea.Batch()
	//return tea.Batch(textinput.Blink)
}

func (m *modelText) setSuggestions(curr string, cmds ...string) {
	commands := make([]string, len(cmds))
	for i, cmd := range cmds {
		commands[i] = curr + "" + strings.TrimPrefix(cmd, curr) + ""
	}

	m.input.SetSuggestions(commands)
}

func (m modelText) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	prev := m.input.Value()

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	curr := m.input.Value()

	//var push bool
	if !strings.HasSuffix(prev, " ") && strings.HasSuffix(curr, " ") {
		//push = true
		//m.debug = "SPACE"
	} else {
		//m.debug = ""
	}

	//var pop bool
	if len(prev) > 0 && len(curr) == 0 {
		//pop = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			//if m.profName == "" {
			//	fmt.Println("Invalid prof name")
			//} else {
			//	fmt.Println("Creating profile with name:", m.profName)
			//}

			return m, tea.Quit
		}
		//case gotReposSuccessMsg:
		//	var suggestions []string
		//	for _, r := range msg {
		//		suggestions = append(suggestions, r.Name)
		//	}
	}

	//suggs, hints := cmdtree.Suggestions(m.args)

	//var args []string
	//for _, arg := range m.args {
	//	args = append(args, strings.TrimLeft(arg, " "))
	//}

	args, err := shlex.Split(m.input.Value())
	if err != nil {
		//args = nil
		m.lex = append(m.lex, "ERR")
	} else {
		if strings.HasSuffix(m.input.Value(), " ") || m.input.Value() == "" {
			args = append(args, "")
		}
	}

	m.lex = args

	//curr, suggs, hint, err := cobrautil.Completion(m.root, append(m.args, m.input.Value()))

	curr, suggs, hint, err := cobrautil.Completion(m.root, args...)

	matched := args
	if len(args) > 0 {
		matched = args[:len(args)-1]
	}

	match := strings.Join(matched, " ")
	if len(match) > 0 {
		match += " "
	}
	//if curr == strings.TrimSpace(m.input.Value()) {
	//	curr = ""
	//}
	//if curr != "" {
	//	curr += " "
	//}
	m.debug = fmt.Sprintf("## cur: %#v, match: %#v . --", curr, match)
	m.setSuggestions(match, suggs...)
	if err == nil {

		//idx := lo.IndexOf(m.args, curr)
		//rem := len(m.args) - idx - 1
		//segs := strings.Split(hint, " ")
		//segs = segs[rem:]
		//hint = strings.Join(segs, " ")
		//m.debug = fmt.Sprint(rem)
		m.input.Placeholder = hint
	} else {
		m.input.Placeholder = "" //err.Error()
	}

	m.input.Placeholder = ""

	//m.input.Placeholder = strings.Join(hints, " ")

	//switch len(m.args) {
	//case 0:
	//	m.setSuggestions(
	//		"node",
	//		"profile",
	//		"link",
	//		"links",
	//		"profiles",
	//	)
	//	m.input.Placeholder = ""
	//case 1:
	//	switch {
	//	case m.args[0] == "node":
	//		m.setSuggestions(
	//			"status",
	//		)
	//		m.input.Placeholder = ""
	//	case m.args[0] == "profile":
	//		m.setSuggestions(
	//			"create",
	//			"delete",
	//			"links",
	//		)
	//		m.input.Placeholder = ""
	//	case m.args[0] == "link":
	//		m.setSuggestions(
	//			"create",
	//			"delete",
	//		)
	//		m.input.Placeholder = ""
	//	default:
	//		m.setSuggestions()
	//		m.input.Placeholder = ""
	//	}
	//case 2:
	//	switch {
	//	case m.args[0] == "node" && m.args[1] == "status":
	//		m.setSuggestions()
	//		m.input.Placeholder = "<name>"
	//	case m.args[0] == "profile" && (m.args[1] == "create" || m.args[1] == "delete"):
	//		m.setSuggestions()
	//		m.input.Placeholder = "<name>"
	//	case m.args[0] == "link" && (m.args[1] == "create" || m.args[1] == "delete"):
	//		m.setSuggestions()
	//		m.input.Placeholder = "<input> <output>"
	//	default:
	//		m.setSuggestions()
	//		m.input.Placeholder = ""
	//	}
	//default:
	//	m.setSuggestions()
	//	m.input.Placeholder = ""
	//}

	//if push && strings.TrimSpace(m.input.Value()) != "" {
	//	m.args = append(m.args, strings.TrimSpace(m.input.Value()))
	//	m.input.SetValue("")
	//	m.debug = "PUSH"
	//	cmds = append(cmds, tea.Tick(0*time.Second, func(t time.Time) tea.Msg {
	//		return t
	//	}))
	//}

	//if strings.HasSuffix(m.input.Value(), " ") && m.input.Value() != " " {
	//
	//	//m.debug += "match"
	//
	//	cmds = append(cmds, tea.Tick(0*time.Second, func(t time.Time) tea.Msg {
	//		return t
	//	}))
	//}

	//if m.input.Value() == "" && len(m.args) == 0 {
	//	m.input.SetValue(" ")
	//}

	//if len(m.args) > 0 && pop {
	//	m.debug = "POP"
	//	//m.debug += "pop"
	//	val := m.args[len(m.args)-1:][0]
	//	m.input.SetValue("" + val)
	//	m.input.SetCursor(len(val) + 1)
	//	m.args = m.args[:len(m.args)-1]
	//
	//	cmds = append(cmds, tea.Tick(0*time.Second, func(t time.Time) tea.Msg {
	//		return t
	//	}))
	//}
	//if m.input.Value() == "" && len(m.args) > 0 {
	//	//m.debug += "unshift"
	//
	//	val := m.args[len(m.args)-1:][0]
	//	m.input.SetValue(" " + val)
	//	m.input.SetCursor(len(val) + 1)
	//	m.args = m.args[:len(m.args)-1]
	//
	//	cmds = append(cmds, tea.Tick(0*time.Second, func(t time.Time) tea.Msg {
	//		return t
	//	}))
	//}

	return m, tea.Batch(cmds...)
}

var frame int

func (m modelText) View() string {
	var v strings.Builder

	v.WriteString("\n>")
	//if len(m.args) > 0 {
	//	v.WriteString(" ")
	//}

	//command := lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Render(strings.Join(m.args, " "))
	//v.WriteString(command)

	//if m.input.Placeholder != "" {
	//	cur := m.input.Cursor
	//	cur.SetChar("@")
	//	cur.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	//	m.input.Cursor = cur
	//}

	//if len(m.args) > 0 {
	//	v.WriteString(" ")
	//}

	var restore bool
	//if m.input.Value() == " " && m.input.Placeholder != "" && len(m.args) > 0 {
	//	//m.input.SetValue("")
	//	//v.WriteString(" ")
	//
	//	restore = true
	//}
	inp := m.input.View()
	if restore {
		//m.input.SetValue(" ")
	}
	//if len(inp) > 0 {
	//	inp = inp[1:]
	//}
	v.WriteString(inp)

	v.WriteString("\n ")

	v.WriteString(strings.Join(m.input.AvailableSuggestions(), "    "))

	//if m.input.Value() == " " && m.input.Placeholder != "" {
	//	v.WriteString(" " + m.input.Placeholder)
	//} else {
	//	inp := lipgloss.NewStyle().Foreground(lipgloss.Color("#756c86")).Render(m.input.View())
	//	v.WriteString(inp)
	//}

	//for i, in := range m.inputs {
	//	l := fmt.Sprintf(" %d [%s]", i, in.View())
	//	_ = l
	//	v.WriteString(l)
	//	//v.WriteString(in.View())
	//}
	v.WriteString("\n\n\n")
	v.WriteString(fmt.Sprintf("input: %#v \nsuggs: %#v \nsugg: %#v \nlex: %#v\n", m.input.Value(), m.input.AvailableSuggestions(), m.input.CurrentSuggestion(), m.lex))
	frame++
	//v.WriteString("\n")

	return v.String()

	//hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render(m.hint)
	//return fmt.Sprintf(
	//	"> %s%s\n%#v\n",
	//	m.textInput.View(), hint, m.debug,
	//)
}
