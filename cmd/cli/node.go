package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"ledctl3/cmd/cli/table"
	"ledctl3/node/event"
	"ledctl3/pkg/uuid"
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
			"NodeConnected",
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

var sinkCmd = &cobra.Command{
	Use: "sink COMMAND",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sink %s\n", args)
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

var sinkConfigCmd = &cobra.Command{
	Use: "config COMMAND",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sink config %s\n", args)
	},
	Args: cobra.MinimumNArgs(1),
}

var sourceConfigSetCmd = &cobra.Command{
	Use: "set [NODE] SOURCE CONFIG",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("source config set %s\n", args)

		b := []byte(args[len(args)-1])

		if !json.Valid(b) {
			panic("invalid json")
		}

		sourceId, err := uuid.Parse(args[len(args)-2])
		if err != nil {
			panic(err)
		}

		c, err := newClient()
		if err != nil {
			panic(err)
		}

		e := event.SetSourceConfigCommand{
			SourceId: sourceId,
			Config:   b,
		}

		err = c.Request(e)
		if err != nil {
			panic(err)
		}

		fmt.Println("Source config updated.")

		time.Sleep(10 * time.Millisecond)
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
			sourceIds := []string{}

			for _, node := range state.Nodes {
				for _, source := range node.Sources {
					names = append(names, node.Name)
					nodeIds = append(nodeIds, node.Id.String())
					sourceIds = append(sourceIds, source.Id.String())
				}
			}

			compls := [][]string{names, nodeIds, sourceIds}

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

				for _, source := range node.Sources {
					compls = append(compls, source.Id.String())
				}
			}

			return compls, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

var sinkConfigSetCmd = &cobra.Command{
	Use: "set [NODE] SINK CONFIG",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sink config set %s\n", args)

		b := []byte(args[len(args)-1])

		if !json.Valid(b) {
			panic("invalid json")
		}

		sinkId, err := uuid.Parse(args[len(args)-2])
		if err != nil {
			panic(err)
		}

		c, err := newClient()
		if err != nil {
			panic(err)
		}

		e := event.SetSinkConfigCommand{
			SinkId: sinkId,
			Config: b,
		}

		err = c.Request(e)
		if err != nil {
			panic(err)
		}

		fmt.Println("Sink config updated.")

		time.Sleep(10 * time.Millisecond)
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
			sinkIds := []string{}

			for _, node := range state.Nodes {
				for _, sink := range node.Sinks {
					names = append(names, node.Name)
					nodeIds = append(nodeIds, node.Id.String())
					sinkIds = append(sinkIds, sink.Id.String())
				}
			}

			compls := [][]string{names, nodeIds, sinkIds}

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

				for _, sink := range node.Sinks {
					compls = append(compls, sink.Id.String())
				}
			}

			return compls, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}
