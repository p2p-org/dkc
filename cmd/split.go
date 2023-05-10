package cmd

import (
	"github.com/p2p-org/dkc/cmd/split"
	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split keystore to distributed wallets",
	Long:  `Allow to split keystore to distributed wallets`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.LogSplit.Info().Msg("starting split function")
		err := split.Run()
		if err != nil {
			utils.LogSplit.Fatal().Err(nil).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)
}
