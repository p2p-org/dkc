package split

func Run() {
	splitRuntime, err := newSplitRuntime()
	if err != nil {
		panic(err)
	}

	err = splitRuntime.createWallets()
	if err != nil {
		panic(err)
	}

	err = splitRuntime.loadWallets()
	if err != nil {
		panic(err)
	}

	err = splitRuntime.saveAccounts()
	if err != nil {
		panic(err)
	}

	err = splitRuntime.checkSignature()
	if err != nil {
		panic(err)
	}
}
