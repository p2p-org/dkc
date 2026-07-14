package cmd

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/p2p-org/dkc/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

var ReleaseVersion string

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "dkc",
	Short: "Dirk Key Converter",
	Long:  `Allow to split and combine keystores and distributed wallets for Dirk`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// pprof is opt-in: the process handles raw private keys, so the
		// profiling endpoint must not be exposed unless explicitly requested
		if viper.GetBool("pprof") {
			go func() {
				utils.Log.Info().Msg("Starting pprof server on :6060")
				utils.Log.Error().Err(http.ListenAndServe("localhost:6060", nil)).Send()
			}()
		}
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
	err := viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		utils.Log.Fatal().Err(err).Send()
	}

	rootCmd.PersistentFlags().Bool("pprof", false, "enable pprof server on localhost:6060")
	err = viper.BindPFlag("pprof", rootCmd.PersistentFlags().Lookup("pprof"))
	if err != nil {
		utils.Log.Fatal().Err(err).Send()
	}

	if err := e2types.InitBLS(); err != nil {
		utils.Log.Fatal().Err(err).Send()
	}
}
func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.Set("version", ReleaseVersion)

	viper.SetEnvPrefix("DKC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		utils.InitLogging(string(rootCmd.Name()))
	}
}
