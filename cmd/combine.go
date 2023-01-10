package cmd

import (
	"github.com/spf13/cobra"
	"github.com/selfuryon/dkc/cmd/combine"
)

var combineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combine distributed wallets to keystore",
	Long:  `Allow to combine distributed wallets to keystore`,
	Run: func(cmd *cobra.Command, args []string) {
    combine.Run()
	},
}

func init() {
	rootCmd.AddCommand(combineCmd)
}

