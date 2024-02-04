package cobrautil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ref https://github.com/golang/debug/pull/8, https://github.com/authzed/zed/pull/188
func ResetSubCommandFlagValues(root *cobra.Command) {
	for _, c := range root.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				f.Value.Set(f.DefValue)
				f.Changed = false
			}
		})
		ResetSubCommandFlagValues(c)
	}
}
