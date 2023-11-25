package convert

import (
	"context"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
)

func Run() error {
	ctx := context.Background()
	utils.Log.Info().Msgf("init input data")
	dataIn, err := input(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain input")
	}

	utils.Log.Info().Msgf("validating input data")
	err = dataIn.validate()
	if err != nil {
		return errors.Wrap(err, "failed to validate input")
	}

	utils.Log.Info().Msgf("processing input data")
	err = process(dataIn)
	if err != nil {
		return errors.Wrap(err, "failed to process")
	}

	return nil
}
