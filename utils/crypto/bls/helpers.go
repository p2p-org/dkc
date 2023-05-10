package bls

import (
	"bytes"
	"encoding/binary"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/p2p-org/dkc/utils"
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

func CompositeKeysCompare(compositeKey []byte, pubkey []byte) error {
	if !bytes.Equal(compositeKey, pubkey) {
		return utils.ErrorPubKeyMatch
	}

	return nil
}

func SignatureCompare(outputSignature []byte, inputSignature []byte) error {
	if !bytes.Equal(outputSignature, inputSignature) {
		return utils.ErrorSignatureMatch
	}
	return nil
}
