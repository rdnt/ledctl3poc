package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCmdTree(t *testing.T) {
	inputs := []string{"input1", "input2", "input3"}
	outputs := []string{"output1", "output2", "output3", "output4"}

	cmdtree := Root().Sub(
		Command("link").Sub(
			Command("create").ArgFunc("input").Args("output"),
			Command("delete").ArgFunc("input").Args("output"),
		),
	)

	//cmdtree := Root().Sub(
	//	Command("link").Sub(
	//		Command("create").Sub(
	//			Args("input", "output").Sub(
	//				Args("output").Sub(),
	//				ArgsFunc(func() []string {
	//					return outputs
	//				}),
	//			),
	//			ArgsFunc(func() []string {
	//				return inputs
	//			}),
	//		),
	//		Command("delete").Sub(
	//			Hint("input", "output").Sub(
	//				Hint("output"),
	//				CommandsFunc(func() []string {
	//					return outputs
	//				}),
	//			),
	//			CommandsFunc(func() []string {
	//				return inputs
	//			}),
	//		),
	//	),
	//)

	args := []string{"link"}
	suggestions, hints := cmdtree.Suggestions(args)
	assert.DeepEqual(t, []string{"create", "delete"}, suggestions)
	assert.DeepEqual(t, []string{}, hints)

	args = []string{"link", "create"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, inputs, suggestions)
	assert.DeepEqual(t, []string{"input", "output"}, hints)

	args = []string{"link", "create", "input"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, outputs, suggestions)
	assert.DeepEqual(t, []string{"output"}, hints)

	args = []string{"link", "create", "input", "output"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, []string{}, suggestions)
	assert.DeepEqual(t, []string{}, hints)

	args = []string{"link", "create", "input", "output", "invalid"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, []string{}, suggestions)
	assert.DeepEqual(t, []string{}, hints)

	args = []string{"link", "delete"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, inputs, suggestions)
	assert.DeepEqual(t, []string{"input", "output"}, hints)

	args = []string{"link", "delete", "input"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, outputs, suggestions)
	assert.DeepEqual(t, []string{"output"}, hints)

	args = []string{"link", "delete", "input", "output"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, []string{}, suggestions)
	assert.DeepEqual(t, []string{}, hints)

	args = []string{"link", "delete", "input", "output", "invalid"}
	suggestions, hints = cmdtree.Suggestions(args)
	assert.DeepEqual(t, []string{}, suggestions)
	assert.DeepEqual(t, []string{}, hints)
}
