package sender

import (
	"encoding/hex"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
)

type Vote struct {
	PubKey    [48]byte
	Signature [96]byte
	EvenType  int
	EventHash []byte
}

func (vote *Vote) Verify(eventHash []byte) {
	blsPubKey, err := bls.PublicKeyFromBytes(vote.PubKey[:])

	if err != nil {
		return
	}
	sig, err := bls.SignatureFromBytes(vote.Signature[:])

	if err != nil {
		return
	}
	if !sig.Verify(blsPubKey, eventHash[:]) {
		return
	}
	println("successfully verified")
}

func AggregatedSignatureAndValidatorBitSet(votes []*Vote, validators []stakingtypes.Validator) ([]byte, uint64, error) {
	signatures := make([][]byte, 0, len(votes))
	voteAddrSet := make(map[string]struct{}, len(votes))

	var votedAddressSet uint64
	for _, v := range votes {
		voteAddrSet[hex.EncodeToString(v.PubKey[:])] = struct{}{}
		signatures = append(signatures, v.Signature[:])
	}

	for idx, valInfo := range validators {
		if _, ok := voteAddrSet[hex.EncodeToString(valInfo.RelayerBlsKey)]; ok {
			votedAddressSet |= 1 << idx
		}
	}

	sigs, err := bls.MultipleSignaturesFromBytes(signatures)

	if err != nil {
		return nil, 0, err
	}

	return bls.AggregateSignatures(sigs).Marshal(), votedAddressSet, nil
}
