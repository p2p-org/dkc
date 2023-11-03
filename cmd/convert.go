package cmd

import (
	"github.com/p2p-org/dkc/cmd/convert"
	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert wallets from different types",
	Long:  `Allow to convert wallets between different types. Supported types are: nd, hd, distributed`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Log.Info().Msgf("starting DKC-%s", viper.Get("version"))
		utils.Log.Info().Msgf("using config file: %s", viper.ConfigFileUsed())
		utils.LogConvert.Info().Msg("starting convert function")
		err := convert.Run()
		if err != nil {
			utils.LogConvert.Fatal().Err(nil).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
