package cobrautil

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
)

func CompletionSuggestions(root *cobra.Command, a ...string) (suggestions []string, err error) {
	compArgs := append([]string{cobra.ShellCompNoDescRequestCmd}, a...)

	out, err := execute(root, compArgs)
	if err != nil {
		return nil, err
	}

	return parseSuggestions(out)
}

const epilogueLen = 2

func parseSuggestions(res string) ([]string, error) {
	lines := strings.Split(res, "\n")
	if len(lines) < epilogueLen {
		return nil, errors.New("invalid completion response")
	}

	directive := lines[len(lines)-epilogueLen : len(lines)-epilogueLen+1][0]
	if directive == ":1" {
		return nil, nil
	}

	return lines[:len(lines)-epilogueLen], nil
}

func execute(root *cobra.Command, args []string) (output string, err error) {
	ResetSubCommandFlagValues(root)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(ioutil.Discard)
	root.SetArgs(args)

	err = root.Execute()

	return buf.String(), err
}
