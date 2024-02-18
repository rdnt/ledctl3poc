package main

import (
	"errors"
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	rootCmd.AddCommand(nodesCmd)

	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeStatusCmd)

	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceConfigCmd)
	sourceConfigCmd.AddCommand(sourceConfigSetCmd)

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
