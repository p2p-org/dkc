package combine

func Run() {
	combineRuntime, err := newCombineRuntime()
	if err != nil {
		panic(err)
	}

	err = combineRuntime.createWalletAndStore()
	if err != nil {
		panic(err)
	}

	err = combineRuntime.storeUpdater()
	if err != nil {
		panic(err)
	}

	err = combineRuntime.checkSignature()
	if err != nil {
		panic(err)
	}
}
