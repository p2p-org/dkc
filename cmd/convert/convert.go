package convert

import "github.com/p2p-org/dkc/utils"

func Run() error {
	utils.LogConvert.Info().Msg("init and validate config")
	convertRuntime, err := newConvertRuntime()
	if err != nil {
		utils.LogConvert.Err(err).Send()
		return err
	}

	err = convertRuntime.validate()
	if err != nil {
		utils.LogConvert.Err(err).Send()
		return err
	}

	utils.LogConvert.Info().Msg("creating wallets")
	err = convertRuntime.createWalletAndStore()
	if err != nil {
		utils.LogConvert.Err(err).Send()
		return err
	}

	utils.LogConvert.Info().Msg("checking signatures")
	err = convertRuntime.checkSignature()
	if err != nil {
		utils.LogConvert.Err(err).Send()
		return err
	}

	return nil
}
