package combine

import "github.com/p2p-org/dkc/utils"

func Run() error {
	utils.LogCombine.Info().Msg("validating config")
	rt, err := newRuntime()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	err = rt.validate()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogCombine.Info().Msg("creating wallets")
	err = rt.createWalletAndStore()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	utils.LogCombine.Info().Msg("updating stores")
	err = rt.storeUpdater()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	utils.LogCombine.Info().Msg("checking signatures")
	err = rt.checkSignature()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	return nil
}
