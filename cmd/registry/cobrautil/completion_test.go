package cobrautil

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

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
		Use:       "gamma-one COMMAND",
		ValidArgs: []string{"argument one", "argument two"},
		Run:       func(cmd *cobra.Command, args []string) {},
	}

	gammaTwo := &cobra.Command{
		Use: "gamma-two COMMAND",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	alpha.AddCommand(beta)
	alpha.AddCommand(gamma)
	gamma.AddCommand(gammaOne, gammaTwo)

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "error")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"beta", "gamma", "completion", "help"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "be")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"beta"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gam")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"gamma"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"gamma-one", "gamma-two"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "ga")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"gamma-one", "gamma-two"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "gamma-")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"gamma-one", "gamma-two"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "gamma-o")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"gamma-one"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "gamma-one")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"gamma-one"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "gamma-onee")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{})
	})

	t.Run("double quote", func(t *testing.T) {
		suggs, err := CompletionSuggestions(alpha, "gamma", "gamma-one", "argument o")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"argument one"})
	})
}

func TestCompletionRoot(t *testing.T) {
	root := cli.Root()

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"help", "link"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "he")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"help"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "li")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"link"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "link")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"link"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "link", "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"create", "delete"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "link", "cr")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"create"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "link", "create")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"create"})
	})

	t.Run("", func(t *testing.T) {
		suggs, err := CompletionSuggestions(root, "link", "create", "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, suggs, []string{"input1", "input2", "input3"})
	})

}

func Test_parseSuggestions(t *testing.T) {
	type args struct {
		res string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Completions",
			args: args{res: `sugg1
sugg2
sugg3
:4
details`},
			want: []string{"sugg1", "sugg2", "sugg3"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.NoError(t, err)
				return err == nil
			},
		},
		{
			name: "Empty completions",
			args: args{res: `:4
details`},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.NoError(t, err)
				return err == nil
			},
		},
		{
			name: "Error directive",
			args: args{res: `:1
details`},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.NoError(t, err)
				return err == nil
			},
		},
		{
			name: "Error directive with completions",
			args: args{res: `sugg1
sugg2
:1
details`},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.NoError(t, err)
				return err == nil
			},
		},
		{
			name: "Invalid response",
			args: args{res: `:4`},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				return err == nil
			},
		},
		{
			name: "Empty response",
			args: args{res: ``},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSuggestions(tt.args.res)
			if !tt.wantErr(t, err, fmt.Sprintf("parseSuggestions(%v)", tt.args.res)) {
				return
			}
			assert.ElementsMatch(t, got, tt.want, "parseSuggestions(%v)", tt.args.res)
		})
	}
}
