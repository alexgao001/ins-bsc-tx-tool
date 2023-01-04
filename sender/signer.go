package sender

import (
	"encoding/hex"
	"fmt"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
	blscmn "github.com/prysmaticlabs/prysm/crypto/bls/common"
)

type Signer struct {
	privkey blscmn.SecretKey
	pubKey  blscmn.PublicKey
}

func NewSigner(privkey []byte) (*Signer, error) {
	privKey, err := blst.SecretKeyFromBytes(privkey)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.PublicKey()
	return &Signer{
		privkey: privKey,
		pubKey:  pubKey,
	}, nil
}

// SignVote sign a vote, data is used to signed to generate the signature
func (signer *Signer) SignVote(vote *Vote, data []byte) error {
	signature := signer.privkey.Sign(data[:])
	vote.EventHash = append(vote.EventHash, data[:]...)
	copy(vote.PubKey[:], signer.pubKey.Marshal()[:])
	copy(vote.Signature[:], signature.Marshal()[:])

	fmt.Printf("sig is %s", hex.EncodeToString(vote.Signature[:]))
	return nil
}
