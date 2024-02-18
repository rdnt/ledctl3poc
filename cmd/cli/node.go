package main

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"ledctl3/cmd/cli/table"
)

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
			"InputIds",
			"OutputsCount",
		}

		nodes := getNodes(state)

		var rows [][]string
		for _, node := range nodes {
			rows = append(rows, []string{
				node.Id,
				node.Name,
				fmt.Sprintf("%t", node.Connected),
				fmt.Sprintf("%d", node.InputsCount),
				fmt.Sprintf("%d", node.OutputsCount),
			})
		}

		t := table.New().
			WithHeaders(headers).
			WithRows(rows).
			WithPadding(4)

		fmt.Println(t.String())
	},
}

var nodeCmd = &cobra.Command{
	Use:  "node NODE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("node %s\n", args)
	},
}

var sourceCmd = &cobra.Command{
	Use: "source COMMAND",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("source %s\n", args)
	},
}

var nodeStatusCmd = &cobra.Command{
	Use: "status NODE",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("node status %s\n", args)
	},
	Args: cobra.ExactArgs(1),
}

var sourceConfigCmd = &cobra.Command{
	Use: "config COMMAND",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("source config %s\n", args)
	},
	Args: cobra.MinimumNArgs(1),
}

var sourceConfigSetCmd = &cobra.Command{
	Use: "set [NODE] SOURCE CONFIG",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("source config set %s\n", args)
	},
	Args: cobra.RangeArgs(2, 3),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			state, err := getState()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			names := []string{}
			nodeIds := []string{}
			driverIds := []string{}

			for _, node := range state.Nodes {
				for _, driver := range node.Drivers {
					names = append(names, node.Name)
					nodeIds = append(nodeIds, node.Id.String())
					driverIds = append(driverIds, driver.Id.String())
				}
			}

			compls := [][]string{names, nodeIds, driverIds}

			return lo.Flatten(compls), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		} else if len(args) == 1 {
			state, err := getState()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			compls := []string{}

			for _, node := range state.Nodes {
				if node.Id.String() != args[0] && node.Name != args[0] {
					continue
				}

				for _, driver := range node.Drivers {
					compls = append(compls, driver.Id.String())
				}
			}

			return compls, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}
