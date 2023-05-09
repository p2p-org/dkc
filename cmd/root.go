package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "dkc",
	Short: "Dirk Key Converter",
	Long:  `Allow to split and combine keystores and distributed wallets for Dirk`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file")

	rootCmd.PersistentFlags().String("log-level", "INFO", "Log Level")
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))

	if err := e2types.InitBLS(); err != nil {
		utils.Log.Fatal().Err(err).Send()
	}
}
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("DKC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		utils.InitLogging()

		msg := fmt.Sprintf("using config file: %s", viper.ConfigFileUsed())
		utils.Log.Info().Msg("starting DKC")
		utils.Log.Info().Msg(msg)
	}
}
