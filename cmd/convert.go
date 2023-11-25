package cmd

import (
	"github.com/p2p-org/dkc/cmd/convert"
	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert from distributed|hd|nd wallet types to distributed or nd wallets types",
	Long:  `Allow to convert wallets types between each other`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.InitLogging(string(cmd.Name()))
		utils.Log.Info().Msgf("starting DKC-%s", viper.Get("version"))
		utils.Log.Info().Msgf("using config file: %s", viper.ConfigFileUsed())
		utils.Log.Info().Msg("starting convert function")
		err := convert.Run()
		if err != nil {
			utils.Log.Fatal().Err(err).Send()
		}
		utils.Log.Info().Msgf("done")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
