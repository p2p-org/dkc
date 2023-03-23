package bls

import (
	"context"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func Check(ctx context.Context) {
	//
}

func Split(ctx context.Context, key []byte, threshold uint32) (
	masterSKs [][]byte,
	masterPKs [][]byte,
) {
	masterSKs, masterPKs = setupMasterKeys(key, threshold)
	return
}

func Recover(ctx context.Context, keys [][]byte, ids []uint64) ([]byte, error) {
	var subIDs []bls.ID
	for _, id := range ids {
		subIDs = append(
			subIDs,
			*newBlsID(id),
		)
	}

	var subSKs []bls.SecretKey

	for _, key := range keys {
		var mk bls.SecretKey
		mk.Deserialize(key)

		subSKs = append(subSKs, mk)
	}

	var rk bls.SecretKey
	if err := rk.Recover(subSKs, subIDs); err != nil {
		panic(err)
	}

	return rk.Serialize(), nil
}
