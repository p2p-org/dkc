package split

import "github.com/p2p-org/dkc/utils"

func Run() error {
	utils.LogSplit.Info().Msg("validating config")
	rt, err := newRuntime()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	err = rt.validate()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("loading wallets")
	err = rt.loadWallets()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("creating wallets")
	err = rt.createWallets()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("saving accounts")
	err = rt.saveAccounts()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogSplit.Info().Msg("checking signatures")
	err = rt.checkSignature()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	return nil
}
