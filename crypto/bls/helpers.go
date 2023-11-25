package bls

import (
	"encoding/binary"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func newBlsID(id uint64) (*bls.ID, error) {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	err := res.SetLittleEndian(buf[:])
	if err != nil {
		return nil, err
	}
	return &res, nil
}
