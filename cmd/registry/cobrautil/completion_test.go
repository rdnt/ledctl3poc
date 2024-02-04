package cobrautil

import (
	"testing"

	"github.com/spf13/cobra"
	assert2 "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"

	"ledctl3/cmd/cli"
)

func TestCompletion(t *testing.T) {
	alpha := &cobra.Command{
		Use: "alpha COMMAND",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	beta := &cobra.Command{
		Use: "beta COMMAND",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	gamma := &cobra.Command{
		Use: "gamma COMMAND",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	gammaOne := &cobra.Command{
		Use: "gamma-one COMMAND",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	gammaTwo := &cobra.Command{
		Use: "gamma-two COMMAND",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	alpha.AddCommand(beta)
	alpha.AddCommand(gamma)
	gamma.AddCommand(gammaOne, gammaTwo)

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "error")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"beta", "gamma", "completion", "help"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "be")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"beta"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gam")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"gamma"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gamma", "")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha gamma")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"gamma-one", "gamma-two"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gamma", "ga")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha gamma")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"gamma-one", "gamma-two"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gamma", "gamma-")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha gamma")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"gamma-one", "gamma-two"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gamma", "gamma-o")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha gamma")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"gamma-one"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gamma", "gamma-one")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha gamma gamma-one")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"gamma-one"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(alpha, "gamma", "gamma-onee")
		assert.NilError(t, err)
		assert.Equal(t, curr, "alpha gamma")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{})
	})
}

func TestCompletionRoot(t *testing.T) {
	root := cli.Root()

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"help", "link"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "he")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"help"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "li")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"link"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "link")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl link")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"link"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "link", "")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl link")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"create", "delete"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "link", "cr")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl link")
		assert.Equal(t, hint, "COMMAND")
		assert2.ElementsMatch(t, sugg, []string{"create"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "link", "create")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl link create")
		assert.Equal(t, hint, "INPUT OUTPUT")
		assert2.ElementsMatch(t, sugg, []string{"create"})
	})

	t.Run("", func(t *testing.T) {
		curr, sugg, hint, err := Completion(root, "link", "create", "")
		assert.NilError(t, err)
		assert.Equal(t, curr, "ledctl link create")
		assert.Equal(t, hint, "INPUT OUTPUT")
		assert2.ElementsMatch(t, sugg, []string{"input1", "input2", "input3"})
	})

}
