package combine

import "github.com/p2p-org/dkc/utils"

func Run() error {
	utils.LogCombine.Info().Msg("validating config")
	combineRuntime, err := newCombineRuntime()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	err = combineRuntime.validate()
	if err != nil {
		utils.LogSplit.Err(err).Send()
		return err
	}

	utils.LogCombine.Info().Msg("creating wallets")
	err = combineRuntime.createWalletAndStore()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	utils.LogCombine.Info().Msg("updating stores")
	err = combineRuntime.storeUpdater()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	utils.LogCombine.Info().Msg("checking signatures")
	err = combineRuntime.checkSignature()
	if err != nil {
		utils.LogCombine.Err(err).Send()
		return err
	}

	return nil
}
