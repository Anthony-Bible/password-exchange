package version

import (
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/buildinfo"
	"github.com/spf13/cobra"
)

// Cmd is exported so tests can drive it directly.
var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version",
	Run: func(c *cobra.Command, args []string) {
		fmt.Fprintln(c.OutOrStdout(), buildinfo.Version)
	},
}

func init() {
	cmd.RootCmd.AddCommand(Cmd)
}
