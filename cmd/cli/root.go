package cli

import (
	"errors"
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(linkCmd)
	linkCmd.AddCommand(linkCreateCmd)
	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:   "ledctl COMMAND",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== root")
	},
}

var linkCmd = &cobra.Command{
	Use:   "link COMMAND",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== link")
	},
}

var linkCreateCmd = &cobra.Command{
	Use:   "create INPUT OUTPUT",
	Short: "",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if !slices.Contains([]string{"input1", "input2", "input3"}, args[0]) {
				return errors.New("invalid input")
			}
		}

		if len(args) > 1 {
			if !slices.Contains([]string{"output1", "output2"}, args[1]) {
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
		if len(args) == 0 && toComplete != "" {
			return []string{"input1", "input2", "input3"}, cobra.ShellCompDirectiveNoFileComp
		} else if len(args) == 1 && slices.Contains([]string{"input1", "input2", "input3"}, args[0]) && toComplete != "" {
			return []string{"output1", "output2"}, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveError
		}
	},
}
