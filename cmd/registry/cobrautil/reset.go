package cobrautil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ResetSubCommandFlagValues resets the flag of all commands, recursively
// ref https://github.com/golang/debug/pull/8
// ref https://github.com/authzed/zed/pull/188
func ResetSubCommandFlagValues(root *cobra.Command) {
	for _, c := range root.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				_ = f.Value.Set(f.DefValue)
				f.Changed = false
			}
		})
		ResetSubCommandFlagValues(c)
	}
}
