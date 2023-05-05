package cmd

import (
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
	Long:  `Allow to split and combine keystores and distributed wallets in Dirk`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Log.Fatal().Err(err)
	}
}

func init() {

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file")

	rootCmd.PersistentFlags().String("log-level", "INFO", "Log Level")
	viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("log-level"))

	utils.InitLogging()

	if err := e2types.InitBLS(); err != nil {
		utils.Log.Fatal().Err(err)
	}
}
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("DKC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.ReadInConfig()

}
