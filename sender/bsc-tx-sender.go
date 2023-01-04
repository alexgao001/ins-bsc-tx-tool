package sender

import (
	"context"
	"encoding/hex"
	"fmt"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"ins-bsc-tx-tool/config"
	"ins-bsc-tx-tool/crosschain"
	"math/big"
)

type BscTxSender struct {
	bscExecutor *BSCExecutor
	insExecutor *InscriptionExecutor
	cfg         *config.Config
}

func NewBscTxSender(bscExecutor *BSCExecutor, insExecutor *InscriptionExecutor, cfg *config.Config) *BscTxSender {
	return &BscTxSender{
		bscExecutor: bscExecutor,
		insExecutor: insExecutor,
		cfg:         cfg,
	}
}

func (s *BscTxSender) Send() (common.Hash, error) {
	nonce, err := s.bscExecutor.client.PendingNonceAt(context.Background(), s.bscExecutor.txSender)
	if err != nil {
		return common.Hash{}, err
	}

	payloadBts := common.Hex2Bytes("746573745061796c6f6164") // "testPayload"
	validators, err := s.insExecutor.QueryLatestValidators()

	aggregatedSig, validatorBitset, err := s.getAggregatedSignatureAndValidatorBitset(payloadBts, validators)

	sequence, err := s.bscExecutor.GetNextSequence()
	if err != nil {
		return common.Hash{}, err
	}

	fmt.Printf("calling smart cotnract with chainid %d, aggregated sig % s, sequence %d, validatorBitset %d, payload %s, nonce %d\n",
		s.cfg.BscConfig.ChannelID,
		hex.EncodeToString(aggregatedSig),
		sequence,
		validatorBitset,
		hex.EncodeToString(payloadBts),
		nonce,
	)
	tx, err := s.CallBuildInSystemContract(s.cfg.BscConfig.ChannelID, aggregatedSig, sequence, validatorBitset, payloadBts, nonce)
	if err != nil {
		return common.Hash{}, err
	}

	fmt.Printf("channelID: %d, sequence: %d, txHash: %s", s.cfg.BscConfig.ChannelID, sequence, tx.String())
	return tx, nil

}

func (s *BscTxSender) getAggregatedSignatureAndValidatorBitset(payload []byte, validators []stakingtypes.Validator) ([]byte, *big.Int, error) {

	var votes []*Vote

	for _, blsPrivKey := range s.cfg.RelayerConfig.BlsPrivateKeys {
		signer, err := NewSigner(common.Hex2Bytes(blsPrivKey))
		if err != nil {
			return nil, nil, err
		}
		encBts, _ := rlp.EncodeToBytes(payload)
		// Hash the rlp-encoded bytes, can eventHash
		eventHash := crypto.Keccak256Hash(encBts).Bytes()
		var vote Vote
		err = signer.SignVote(&vote, eventHash)
		if err != nil {
			return nil, nil, err
		}
		votes = append(votes, &vote)
	}

	aggregatedSignature, votedAddressSet, err := AggregatedSignatureAndValidatorBitSet(votes, validators)
	if err != nil {
		return nil, nil, err
	}
	return aggregatedSignature, big.NewInt(int64(votedAddressSet)), nil
}

func (s *BscTxSender) CallBuildInSystemContract(channelID int8, blsSignature []byte, sequence uint64, validatorSet *big.Int,
	msgBytes []byte, nonce uint64) (common.Hash, error) {

	txOpts, err := s.bscExecutor.getTransactor(nonce)
	if err != nil {
		return common.Hash{}, err
	}

	crossChainInstance, err := crosschain.NewCrosschain(crossChainContractAddr, s.bscExecutor.client)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := crossChainInstance.HandlePackage(txOpts, msgBytes, blsSignature, validatorSet, sequence, uint8(channelID))
	if err != nil {
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}
