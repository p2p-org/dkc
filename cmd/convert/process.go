package convert

import (
	"fmt"
	"sync"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func process(data *dataIn) error {
	//Init Stores
	iStore := data.InputS

	oStore := data.OutputS

	// Create Output Store
	utils.Log.Info().Msgf(fmt.Sprintf("creating output store %s", oStore.GetPath()))
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
			// Get Private Key From Account
			utils.Log.Info().Msgf("getting private key for account %s from wallet %s", aName, wName)
			pk, err := iStore.GetPK(wName, aName)
			if err != nil {
				aPKMap <- a{name: aName, pk: []byte{}, err: err, wName: wName}
			}
			aPKMap <- a{name: aName, pk: pk, err: nil, wName: wName}
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
			return err
		}
		// Save Private Key To Output Store
		utils.Log.Info().Msgf("converting and saving private key for account %s to output wallet %s", a.name, a.wName)
		err = oStore.SavePKToWallet(a.wName, a.pk, a.name)
		if err != nil {
			return errors.Wrap(err, "failed to save private key to output wallet")
		}
	}

	return nil
}
