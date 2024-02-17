package main

import (
	"bytes"
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
	root            *cobra.Command
	input           textinput.Model
	parsedShellArgs []string
}

func (m modelText) Init() tea.Cmd {
	return tea.Batch()
	//return tea.Batch(textinput.Blink)
}

func (m *modelText) setSuggestions(curr string, cmds ...string) {
	for i, cmd := range cmds {
		//commands[i] = cmd

		//for i, ch := range curr {
		//	for j, cmd := range cmds {
		//		var offset int
		//		if i <= len(cmd) {
		//			//cmd = curr + strings.TrimPrefix(cmd, curr)
		//			if ch == '"' || ch == '\'' {
		//				var cmdr []rune
		//				cmdr = append(cmdr, []rune(cmd)[:i]...)
		//				cmdr = append(cmdr, ch)
		//				cmdr = append(cmdr, []rune(cmd)[i:]...)
		//				cmds[j] = string(cmdr)
		//				offset++
		//			}
		//		}
		//	}
		//}

		cmds[i] = curr + "" + strings.TrimPrefix(cmd, "") + ""
	}

	m.input.SetSuggestions(cmds)
}

func (m modelText) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			args, err := shlex.Split(m.input.Value())
			if err != nil {
				panic(err)
			}

			cobrautil.ResetSubCommandFlagValues(m.root)

			buf := new(bytes.Buffer)
			m.root.SetOut(buf)
			m.root.SetErr(buf)
			m.root.SetArgs(args)

			err = m.root.Execute()
			if err != nil {
				panic(err)
			}

			fmt.Print(buf.String())

			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	args, _ := shlex.Split(m.input.Value())

	if strings.HasSuffix(m.input.Value(), " ") || m.input.Value() == "" {
		args = append(args, "")
	}

	//m.parsedShellArgs = args

	suggs, _ := cobrautil.CompletionSuggestions(m.root, args...)

	//matched := args
	//if len(args) > 0 {
	//	matched = args[:len(args)-1]
	//}

	val := m.input.Value()
	idx := strings.LastIndex(val, " ")
	if idx != -1 {
		val = val[:idx]
	} else {
		val = ""
	}

	//match := strings.Join(matched, " ")
	if len(val) > 0 {
		val += " "
	}

	m.setSuggestions(val, suggs...)

	//if len(m.parsedShellArgs) > 0 && m.parsedShellArgs[len(m.parsedShellArgs)-1] != "" {
	//	m.setSuggestions(match, suggs...)
	//} else {
	//	m.setSuggestions(match, suggs...)
	//}

	return m, tea.Batch(cmds...)
}

var frame int

func (m modelText) View() string {
	var v strings.Builder

	v.WriteString(">")
	//if len(m.args) > 0 {
	v.WriteString(" ")
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

	v.WriteString("\n  ")

	var suggs []string
	suggs = m.input.AvailableSuggestions()
	//for _, sug := range m.input.AvailableSuggestions() {
	//	segs := strings.Split(sug, " ")
	//	if len(segs) > 0 && segs[len(segs)-1] != "" {
	//		suggs = append(suggs, segs[len(segs)-1])
	//	}
	//}
	v.WriteString(strings.Join(suggs, "  "))

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
	v.WriteString("\n\n")
	v.WriteString(fmt.Sprintf("input: %#v \nsuggs: %#v \nselected: %#v \nhint: %#v \nparsedShellArgs: %#v\n", m.input.Value(), m.input.AvailableSuggestions(), m.input.CurrentSuggestion(), m.input.Placeholder, m.parsedShellArgs))
	frame++
	//v.WriteString("\n")

	return v.String()

	//hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render(m.hint)
	//return fmt.Sprintf(
	//	"> %s%s\n%#v\n",
	//	m.textInput.View(), hint, m.debug,
	//)
}
