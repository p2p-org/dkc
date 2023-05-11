package cmd

import (
	"github.com/p2p-org/dkc/cmd/combine"
	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var combineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combine distributed wallets to keystore",
	Long:  `Allow to combine distributed wallets to keystore`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Log.Info().Msgf("starting DKC-%s", viper.Get("version"))
		utils.Log.Info().Msgf("using config file: %s", viper.ConfigFileUsed())
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
