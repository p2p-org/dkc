package bls

import (
	"encoding/binary"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func newBlsID(id uint64) *bls.ID {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	err := res.SetLittleEndian(buf[:])
	if err != nil {
		panic(err)
	}
	return &res
}
