package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"
)

func execute(root *cobra.Command, args []string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func main() {
	//rootCmd.SetOut()

	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(linkCmd)
	linkCmd.AddCommand(createCmd)

	//if err := rootCmd.Execute(); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//os.Exit(0)

	//fmt.Println(rootCmd.Commands())
	//for _, c := range rootCmd.Commands() {
	//	//fmt.Println(c.Name())
	//	if c.Name() == "link" {
	//		for _, c := range c.Commands() {
	//			//fmt.Println(c.Name())
	//			if c.Name() == "create" {
	//				//fmt.Println(c.ValidArgsFunction())
	//			}
	//		}
	//	}
	//}

	cmd := ""

	args, err := shlex.Split(cmd)
	if err != nil {
		panic(err)
	}

	compArgs := append([]string{cobra.ShellCompNoDescRequestCmd}, args...)
	if len(args) == 0 {
		compArgs = append(compArgs, "")
	}

	c, out, err := execute(rootCmd, compArgs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	footerLines := 3
	opts := strings.Split(out, "\n")
	opts = opts[:len(opts)-footerLines]

	fmt.Printf("SUGG:\n%v\n", opts)

	if rootCmd.TraverseChildren {
		c, _, err = rootCmd.Traverse(args)
	} else {
		c, _, err = rootCmd.Find(args)
	}

	fmt.Printf("NAME:\n%s\n", c.Name())
	fmt.Printf("HINT:\n%s\n", strings.TrimPrefix(c.Use, c.Name()+" "))

	c, out, err = execute(rootCmd, args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//fmt.Printf("OUT\n%s\n", out)

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

var createCmd = &cobra.Command{
	Use:   "create INPUT OUTPUT",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== create")
	},
	Args: cobra.RangeArgs(1, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"input1", "input2", "input3"}, cobra.ShellCompDirectiveNoFileComp
		} else if len(args) == 1 {
			return []string{"output1", "output2"}, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	},

	//ValidArgs: []string{
	//	"link1",
	//	"link2",
	//	"link3",
	//},
}
