package cli

import (
	"fmt"

	"github.com/mydehq/autotitle"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("autotitle %s\n", autotitle.Version())
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
