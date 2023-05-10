package cmd

import (
	"github.com/p2p-org/dkc/cmd/combine"
	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
)

var combineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combine distributed wallets to keystore",
	Long:  `Allow to combine distributed wallets to keystore`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.LogCombine.Info().Msg("starting combine function")
		err := combine.Run()
		if err != nil {
			utils.LogCombine.Fatal().Err(nil).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(combineCmd)
}
