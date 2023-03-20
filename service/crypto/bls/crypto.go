package bls

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func newBlsID(id uint64) *bls.ID {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	if err := res.SetLittleEndian(buf[:]); err != nil {
		panic(err)
	}
	return &res
}

func Recover(ctx context.Context, keys [][]byte, ids []uint64) (string, error) {
	bls.Init(bls.BLS12_381)
	bls.SetETHmode(bls.EthModeDraft07)

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
	}
	fmt.Printf("Recovered key=%v\n", rk.SerializeToHexStr())
	return "", nil
}
