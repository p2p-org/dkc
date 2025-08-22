package convert

import (
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
	accountsList, walletsList, err := iStore.GetAccounts()
	if err != nil {
		return errors.Wrap(err, "failed to get AccountsWallets map")
	}

	// Create Wallets In Output Store First
	utils.Log.Info().Msgf("creating wallets in output store")
	for _, w := range walletsList {
		utils.Log.Info().Msgf("creating wallet %s in %s", w, oStore.GetPath())
		err := oStore.CreateWallet(w)
		if err != nil {
			return errors.Wrap(err, "failed to create wallet")
		}
	}

	// Converting Accounts - Full pipeline per account in parallel
	utils.Log.Info().Msgf("converting accounts with parallel GetPrivateKey -> SavePrivateKey pipeline")

	// Use errgroup to handle errors from parallel account processing
	accountGroup := errgroup.Group{}

	// Process each account in parallel: GetPrivateKey -> SavePrivateKey
	for _, acc := range accountsList {
		// Capture loop variables
		aName := acc.Name
		wName := acc.WName

		accountGroup.Go(func() error {
			// Get Private Key From Account
			utils.Log.Info().Msgf("getting private key for account %s from wallet %s", aName, wName)
			pk, err := iStore.GetPrivateKey(wName, aName)
			if err != nil {
				utils.Log.Error().Err(err).Msgf("failed to get private key for account %s from wallet %s", aName, wName)
				return errors.Wrap(err, "failed to get private key")
			}

			utils.Log.Info().Msgf("got private key for account %s from wallet %s", aName, wName)

			// Save Private Key To Output Store
			utils.Log.Info().Msgf("converting and saving private key for account %s to output wallet %s", aName, wName)
			err = oStore.SavePrivateKey(wName, aName, pk)
			if err != nil {
				utils.Log.Error().Err(err).Msgf("failed to save private key for account %s to wallet %s", aName, wName)
				return errors.Wrap(err, "failed to save private key to output wallet")
			}

			utils.Log.Info().Msgf("✅ successfully converted account %s in wallet %s", aName, wName)
			return nil
		})
	}

	// Wait for all account conversions to complete
	if err := accountGroup.Wait(); err != nil {
		return err
	}

	return nil
}
