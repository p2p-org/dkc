package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of DKC",
	Long:  `Current release version of DKC`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printDKCVersion()
		return nil
	},
}

func printDKCVersion() {
	fmt.Println(viper.Get("version"))
}
