package cmd

import (
	"fmt"
	"os"
  "strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

var (
	cfgFile     string
)

var rootCmd = &cobra.Command{
	Use:   "dkc",
	Short: "Dirk Key Converter",
	Long:  `Allow to split and combine keystores and distributed wallets in Dirk`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello")
    fmt.Printf("Viper: %+v\n", viper.GetStringMap("peers"))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	if err := e2types.InitBLS(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file")

	rootCmd.PersistentFlags().String("keystore-dir", "./keystores", "Directory with keystores")
  viper.BindPFlag("keystoreDir", rootCmd.PersistentFlags().Lookup("keystore-dir"))

	rootCmd.PersistentFlags().String("wallet-dir", "./wallets", "Directory with dirk wallets")
  viper.BindPFlag("walletDir", rootCmd.PersistentFlags().Lookup("wallet-dir"))
}
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("DKC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
