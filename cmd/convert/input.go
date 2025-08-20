package convert

import (
	"context"

	store "github.com/p2p-org/dkc/store"
	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type dataIn struct {
	Ctx         context.Context
	InputStore  store.InputStore
	OutputStore store.OutputStore
}

func (d *dataIn) validate() error {
	utils.Log.Debug().Msgf("comparing input and output paths")
	utils.Log.Debug().Msgf("input store: %s, output store: %s", d.InputStore.GetPath(), d.OutputStore.GetPath())
	if d.InputStore.GetPath() == d.OutputStore.GetPath() {
		return errors.New("paths are the same")
	}

	utils.Log.Debug().Msgf("comparing input and output types")
	utils.Log.Debug().Msgf("input store: %s, output store: %s", d.InputStore.GetType(), d.OutputStore.GetType())
	if d.InputStore.GetType() == d.OutputStore.GetType() {
		return errors.New("wallets types are the same")
	}

	return nil
}

func input(ctx context.Context) (*dataIn, error) {
	var err error
	data := &dataIn{}

	data.Ctx = ctx
	//Parse Input Config
	utils.Log.Debug().Msgf("init %s as input store", viper.GetString("input.wallet.type"))
	data.InputStore, err = store.InputStoreInit(ctx, viper.GetString("input.wallet.type"))
	if err != nil {
		return nil, err
	}

	//Parse Output Config
	utils.Log.Debug().Msgf("init %s as output store", viper.GetString("output.wallet.type"))
	data.OutputStore, err = store.OutputStoreInit(ctx, viper.GetString("output.wallet.type"))
	if err != nil {
		return nil, err
	}

	return data, nil
}
