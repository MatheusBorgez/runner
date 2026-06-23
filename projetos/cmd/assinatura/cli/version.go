package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version é sobrescrita em build via -ldflags "-X github.com/kyriosdata/runner/cmd/assinatura/cli.version=<tag>"
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão atual do CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("assinatura %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
