package main

type Cmd struct {
	name string
	args []arg
	sub  []Cmd
}

type arg struct {
	name    string
	options func() []string
}

func (c Cmd) Sub(cmds ...Cmd) Cmd {
	c.sub = append(c.sub, cmds...)
	return c
}

func (c Cmd) Arg(a string) Cmd {
	c.args = append(c.args, arg{
		name: a,
	})
	return c
}

func (c Cmd) ArgFunc(arg string, options []string) Cmd {
	c.args = append(c.args)
	return c
}

func Command(cmd string) Cmd {
	return Cmd{name: cmd}
}

//func Args(args ...string) Cmd {
//	return Cmd{argsf: func() []string {
//		return args
//	}}
//}
//
//func ArgsFunc(args func() []string) Cmd {
//	return Cmd{argsf: args}
//}

func Root() Cmd {
	return Cmd{}
}

func (c Cmd) Suggestions(args []string) ([]string, []string) {
	var suggestions []string
	var hints []string

	//curr := c

	suggestions = []string{}
	hints = []string{}

	//if len(args) == 0 {
	//	for _, name := range curr.sub {
	//		for _, c := range name.name {
	//			if c != "" {
	//				suggestions = append(suggestions, c)
	//			}
	//		}
	//
	//		for _, h := range name.args {
	//			if h != "" {
	//				hints = append(hints, h)
	//			}
	//		}
	//	}
	//
	//	return suggestions, hints
	//}
	//
	//for i := 0; i < len(args); i++ {
	//	var sub Cmd
	//	var found bool
	//	for _, name := range curr.sub {
	//		for _, h := range name.args {
	//			if i > 0 && h != "" {
	//				sub = name
	//				found = true
	//			}
	//		}
	//
	//		for _, c := range name.name {
	//			if c == args[i] {
	//				sub = name
	//				found = true
	//				//break
	//			}
	//		}
	//	}
	//
	//	if !found {
	//		break
	//	}
	//
	//	curr = sub
	//
	//	suggestions = []string{}
	//	hints = []string{}
	//	for _, name := range curr.sub {
	//		for _, c := range name.name {
	//			if c != "" {
	//				suggestions = append(suggestions, c)
	//			}
	//		}
	//
	//		for _, h := range name.args {
	//			if h != "" {
	//				hints = append(hints, h)
	//			}
	//		}
	//	}
	//}

	return suggestions, hints
}
