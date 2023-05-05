package cmd

import (
	"github.com/p2p-org/dkc/cmd/combine"
	"github.com/spf13/cobra"
)

var combineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combine distributed wallets to keystore",
	Long:  `Allow to combine distributed wallets to keystore`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := combine.Run()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(combineCmd)
}
