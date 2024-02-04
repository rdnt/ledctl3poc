package cobrautil

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
)

func Completion(root *cobra.Command, a ...string) (name string, sugg []string, hint string, err error) {

	//fmt.Printf("ARGS %#v %#v\n", root.Use, a)

	compArgs := append([]string{cobra.ShellCompNoDescRequestCmd}, a...)
	//compArgs = append(compArgs, "")

	// Output from stderr must be ignored by the completion script.
	c, out, err := execute(root, compArgs)
	if err != nil {
		return "", nil, "", err
	}

	footerLines := 2
	sugg = strings.Split(out, "\n")
	directive := sugg[len(sugg)-footerLines : len(sugg)-footerLines+1][0]
	if directive != ":1" {
		sugg = sugg[:len(sugg)-footerLines]
	} else {
		return "", nil, "", errors.New("invalid")
	}

	if len(sugg) == 0 {
		// try without appended ""

		//compArgs := append([]string{cobra.ShellCompNoDescRequestCmd}, a...)
		//
		//// Output from stderr must be ignored by the completion script.
		//c, out, err = execute(root, compArgs)
		//if err != nil {
		//	return "", nil, "", err
		//}
		//
		//footerLines := 2
		//sugg = strings.Split(out, "\n")
		//directive := sugg[len(sugg)-footerLines : len(sugg)-footerLines+1][0]
		//if directive != ":1" {
		//	sugg = sugg[:len(sugg)-footerLines]
		//} else {
		//	return "", nil, "", errors.New("invalid")
		//}
	}

	if root.TraverseChildren {
		c, _, err = root.Traverse(a)
	} else {
		c, _, err = root.Find(a)
	}
	hint = strings.TrimPrefix(c.Use, c.Name()+" ")

	//fmt.Printf("NAME:\n%s\n", c.Name())
	//fmt.Printf("HINT:\n%s\n", strings.TrimLeft(c.Use, c.Name()+" "))

	return strings.TrimPrefix(c.CommandPath(), " "), sugg, hint, nil
}

func execute(root *cobra.Command, args []string) (c *cobra.Command, output string, err error) {
	ResetSubCommandFlagValues(root)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(ioutil.Discard)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
