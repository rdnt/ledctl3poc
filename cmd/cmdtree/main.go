package main

import "fmt"

type command struct {
	cmd  string
	hint []string
	sub  []command
}

func (c command) Sub(cmds ...command) command {
	c.sub = cmds
	return c
}

func Command(cmd string) command {
	return command{cmd: cmd}
}

//func CommandsFunc(cmdsFunc func() []string) command {
//	return command{cmd: cmdsFunc()}
//}

func Root() command {
	return command{}
}

func Hint(hints ...string) command {
	return command{hint: hints}
}

func (c command) Suggestions(args []string) ([]string, []string) {
	var suggestions = []string{}
	var hints = []string{}

	curr := c

	suggestions = []string{}
	hints = []string{}

	if len(args) == 0 {
		for _, cmd := range curr.sub {
			for _, c := range cmd.cmd {
				if c != "" {
					suggestions = append(suggestions, c)
				}
			}

			for _, h := range cmd.hint {
				if h != "" {
					hints = append(hints, fmt.Sprintf("<%s>", h))
				}
			}
		}

		return suggestions, hints
	}

	for i := 0; i < len(args); i++ {
		var sub command
		var found bool
		for _, cmd := range curr.sub {
			for _, h := range cmd.hint {
				if i > 0 && h != "" {
					sub = cmd
					found = true
				}
			}

			for _, c := range cmd.cmd {
				if c == args[i] {
					sub = cmd
					found = true
					//break
				}
			}
		}

		if !found {
			break
		}

		curr = sub

		suggestions = []string{}
		hints = []string{}
		for _, cmd := range curr.sub {
			for _, c := range cmd.cmd {
				if c != "" {
					suggestions = append(suggestions, c)
				}
			}

			for _, h := range cmd.hint {
				if h != "" {
					hints = append(hints, fmt.Sprintf("<%s>", h))
				}
			}
		}
	}

	return suggestions, hints
}

func main() {

	nodes := []string{"potato", "tomato", "teeth"}
	inputs := []string{"input1", "input2", "input3"}
	outputs := []string{"output1", "output2", "output3", "output4"}

	cmdtree := Root().Sub(
		Hint("command"),
		Command("node").Sub(
			Command("status").
				Sub(
					Hint("name"),
					CommandsFunc(func() []string {
						return nodes
					}),
				),
		),
		Command("profiles"),
		Command("profile").Sub(
			Command("create").Sub(Hint("name")),
			Command("delete").Sub(Hint("name")),
			Command("links"),
		),

		Command("links"),
		Command("link").Sub(
			Command("create").Sub(
				Hint("input", "output").Sub(
					Hint("output"),
					CommandsFunc(func() []string {
						return outputs
					}),
				),
				CommandsFunc(func() []string {
					return inputs
				}),
			),
			Command("delete").Sub(
				Hint("input").Sub(Hint("output")),
				Hint("output"),
			),
		),
	)

	args := []string{"link", "create"}
	suggestions, hints := cmdtree.Suggestions(args)
	fmt.Printf("sugg: %v, hints %v, args %v\n", suggestions, hints, args)

	args = []string{"link", "create", "1"}
	suggestions, hints = cmdtree.Suggestions(args)
	fmt.Printf("sugg: %v, hints %v, args %v\n", suggestions, hints, args)

	args = []string{"link", "create", "2", "2"}
	suggestions, hints = cmdtree.Suggestions(args)
	fmt.Printf("sugg: %v, hints %v, args %v\n", suggestions, hints, args)

	args = []string{"link", "create", "3", "3", "3"}
	suggestions, hints = cmdtree.Suggestions(args)
	fmt.Printf("sugg: %v, hints %v, args %v\n", suggestions, hints, args)
	// > <command>
}
