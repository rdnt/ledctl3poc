package main

import (
	"errors"
	"fmt"
	"slices"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"ledctl3/cmd/cli/table"
)

func Root() *cobra.Command {
	rootCmd.AddCommand(nodesCmd)

	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(linkCmd)

	linkCmd.AddCommand(linkCreateCmd)
	linkCmd.AddCommand(linkDeleteCmd)

	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:   "ledctl COMMAND",
	Short: "",
	Long:  "",
	//DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== root")
	},
}

var helpCmd = &cobra.Command{
	Use: "help",
	//DisableFlagParsing: true,
}

var nodesCmd = &cobra.Command{
	Use: "nodes",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := getState()
		if err != nil {
			panic(err)
		}

		headers := []string{
			"Id",
			"Name",
			"Connected",
			"Inputs",
			"Outputs",
		}

		var rows [][]string

		if len(state.Nodes) == 0 {
			rows = append(rows, []string{"(empty)"})
		}

		ids := lo.Keys(state.Nodes)
		slices.Sort(ids)

		for _, id := range ids {
			node := state.Nodes[id]
			id := node.Id.String()
			name := node.Name
			if name == "" {
				name = "-"
			}
			connected := fmt.Sprintf("%t", node.Connected)
			inputs := fmt.Sprintf("%d", len(node.Inputs))
			outputs := fmt.Sprintf("%d", len(node.Outputs))

			rows = append(rows, []string{id, name, connected, inputs, outputs})
		}

		t := table.New().
			WithHeaders(headers).
			WithRows(rows).
			WithPadding(4)

		fmt.Println(t.String())
	},
}

var linkCmd = &cobra.Command{
	Use:   "link COMMAND",
	Short: "",
	Long:  "",
	//DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== link")
	},
}

var linkCreateCmd = &cobra.Command{
	Use:   "create INPUT OUTPUT",
	Short: "",
	Long:  "",
	//DisableFlagParsing: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if !slices.Contains([]string{"input1", "input2", "input3", "input four"}, args[0]) {
				return errors.New("invalid input")
			}
		}

		if len(args) > 1 {
			if !slices.Contains([]string{"output1", "output2", "output three"}, args[1]) {
				return errors.New("invalid output")
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== create")
	},
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"input1", "input2", "input3", "input four"}, cobra.ShellCompDirectiveNoFileComp
		} else if len(args) == 1 && slices.Contains([]string{"input1", "input2", "input3"}, args[0]) {
			return []string{"output1", "output2", "output three"}, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveError
		}
	},
}

var linkDeleteCmd = &cobra.Command{
	Use:   "delete INPUT OUTPUT",
	Short: "",
	Long:  "",
	//DisableFlagParsing: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if !slices.Contains([]string{"input1", "input2", "input3", "input four"}, args[0]) {
				return errors.New("invalid input")
			}
		}

		if len(args) > 1 {
			if !slices.Contains([]string{"output1", "output2", "output three"}, args[1]) {
				return errors.New("invalid output")
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== create")
	},
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"input1", "input2", "input3", "input four"}, cobra.ShellCompDirectiveNoFileComp
		} else if len(args) == 1 && slices.Contains([]string{"input1", "input2", "input3"}, args[0]) {
			return []string{"output1", "output2", "output three"}, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveError
		}
	},
}
