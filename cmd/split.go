package cmd

import (
	"github.com/p2p-org/dkc/cmd/split"
	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split keystore to distributed wallets",
	Long:  `Allow to split keystore to distributed wallets`,
	Run: func(cmd *cobra.Command, args []string) {
		split.Run()
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)
}
