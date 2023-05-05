package combine

func Run() error {
	combineRuntime, err := newCombineRuntime()
	if err != nil {
		return err
	}

	err = combineRuntime.createWalletAndStore()
	if err != nil {
		return err
	}

	err = combineRuntime.storeUpdater()
	if err != nil {
		return err
	}

	err = combineRuntime.checkSignature()
	if err != nil {
		return err
	}

	return nil
}
