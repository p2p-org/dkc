package bls

import (
	"encoding/binary"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
)

func newBlsID(id uint64) (*bls.ID, error) {
	// A zero ID is invalid: evaluating the sharing polynomial at x=0 yields
	// the master secret itself, so the "shard" would be the original key
	if id == 0 {
		return nil, errors.New("participant id must not be zero")
	}

	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	err := res.SetLittleEndian(buf[:])
	if err != nil {
		return nil, err
	}
	return &res, nil
}
