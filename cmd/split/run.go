package split

func Run() error {
	splitRuntime, err := newSplitRuntime()
	if err != nil {
		return err
	}

	err = splitRuntime.createWallets()
	if err != nil {
		return err
	}

	err = splitRuntime.loadWallets()
	if err != nil {
		return err
	}

	err = splitRuntime.saveAccounts()
	if err != nil {
		return err
	}

	err = splitRuntime.checkSignature()
	if err != nil {
		return err
	}

	return nil
}
