package convert

import (
	"sync"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func process(data *dataIn) error {
	//Init Stores
	iStore := data.InputStore

	oStore := data.OutputStore

	// Create Output Store
	utils.Log.Info().Msgf("creating output store %s", oStore.GetPath())
	err := oStore.Create()
	if err != nil {
		return errors.Wrap(err, "failed to create output store")
	}

	// Get Accounts Wallets Map
	utils.Log.Info().Msgf("getting AccountsWallets map for store %s", iStore.GetPath())
	accountsList, walletsList, err := iStore.GetWalletsAccountsMap()
	if err != nil {
		return errors.Wrap(err, "failed to get AccountsWallets map")
	}

	// Converting Wallets
	utils.Log.Info().Msgf("converting accounts")

	// Init channel struct for account
	type a struct {
		name  string
		pk    []byte
		err   error
		wName string
	}
	// Creating channel for wallets
	var wgA sync.WaitGroup
	aPKMap := make(chan a, len(accountsList))

	// Iterates over accounts in separate goroutines
	for _, acc := range accountsList {
		wgA.Add(1)
		go func(aName string, wName string) {
			defer wgA.Done()
			var aPK a
			// Get Private Key From Account
			utils.Log.Info().Msgf("getting private key for account %s from wallet %s", aName, wName)
			pk, err := iStore.GetPrivateKey(wName, aName)
			if err != nil {
				utils.Log.Error().Err(err).Msgf("failed to get private key for account %s from wallet %s", aName, wName)
				aPK = a{name: aName, pk: []byte{}, err: err, wName: wName}
			} else {
				utils.Log.Info().Msgf("got private key for account %s from wallet %s", aName, wName)
				aPK = a{name: aName, pk: pk, err: nil, wName: wName}
			}
			aPKMap <- aPK
		}(acc.Name, acc.WName)
	}
	wgA.Wait()
	close(aPKMap)

	// Create Wallet In Output Store
	group := errgroup.Group{}
	group.Go(func() error {
		for _, w := range walletsList {
			utils.Log.Info().Msgf("creating wallet %s in %s", w, oStore.GetPath())
			err := oStore.CreateWallet(w)
			if err != nil {
				return errors.Wrap(err, "failed to create wallet")
			}
		}
		return nil
	})
	// Return only first error
	if err := group.Wait(); err != nil {
		return err
	}

	//We can get private keys in different goroutines, but we need to save save them one by one
	for a := range aPKMap {

		if a.err != nil {
			return a.err
		}
		// Save Private Key To Output Store
		utils.Log.Info().Msgf("converting and saving private key for account %s to output wallet %s", a.name, a.wName)
		err = oStore.SavePrivateKey(a.wName, a.name, a.pk)
		if err != nil {
			return errors.Wrap(err, "failed to save private key to output wallet")
		}
	}

	return nil
}
