package split

import "github.com/p2p-org/dkc/utils"

func Run() error {
	utils.LogSplit.Info().Msg("validating config")
	splitRuntime, err := newSplitRuntime()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("creating wallets")
	err = splitRuntime.createWallets()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("loading wallets")
	err = splitRuntime.loadWallets()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("saving accounts")
	err = splitRuntime.saveAccounts()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("checking signatures")
	err = splitRuntime.checkSignature()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	return nil
}
