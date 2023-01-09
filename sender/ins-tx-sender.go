package sender

import (
	"context"
	"fmt"
	clitx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/willf/bitset"
	"ins-bsc-tx-tool/config"
	"time"
)

type InsTxSender struct {
	executor *InscriptionExecutor
	cfg      *config.Config
}

func NewInsTxSender(e *InscriptionExecutor, cfg *config.Config) *InsTxSender {
	return &InsTxSender{
		executor: e,
		cfg:      cfg,
	}
}

func (s *InsTxSender) Send() (string, error) {

	txConfig := authtx.NewTxConfig(s.executor.cdc, authtx.DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()

	payloadBts := common.Hex2Bytes("746573745061796c6f6164") // "testPayload"
	privKey, _ := HexToEthSecp256k1PrivKey(s.cfg.RelayerConfig.RelayerInsPrivateKey)
	validatorAddress := privKey.PubKey().Address().String()

	//fromAddress, _ := sdktypes.AccAddressFromHexUnsafe(validatorAddress)
	//toAddress, _ := sdktypes.AccAddressFromHexUnsafe(s.cfg.InsConfig.ToAddress)
	//
	//msg := banktypes.NewMsgSend(fromAddress, toAddress, sdktypes.NewCoins(sdktypes.NewInt64Coin("bnb", 1)))
	//err := txBuilder.SetMsgs(msg)

	validators, err := s.executor.QueryLatestValidators()
	if err != nil {
		return "", nil
	}
	aggregatedSig, validatorBitset, err := s.getAggregatedSignatureAndValidatorBitset(payloadBts, validators)

	//Todo fix
	_, err = s.executor.GetNextOracleSequence()
	if err != nil {
		return "", nil
	}

	msgClaim := &oracletypes.MsgClaim{}
	msgClaim.FromAddress = validatorAddress
	msgClaim.Payload = payloadBts
	msgClaim.VoteAddressSet = validatorBitset.Bytes()
	msgClaim.VoteAddressSet = append(msgClaim.VoteAddressSet, 0, 0, 0)
	msgClaim.Sequence = 1
	msgClaim.AggSignature = aggregatedSig
	msgClaim.DestChainId = s.cfg.InsConfig.DestChainId
	msgClaim.SrcChainId = s.cfg.InsConfig.SrcChainId
	msgClaim.Timestamp = uint64(time.Now().Unix())
	err = txBuilder.SetMsgs(msgClaim)

	if err != nil {
		panic(err)
	}
	txBuilder.SetGasLimit(210000)

	acct, err := s.executor.GetAccount(validatorAddress)
	if err != nil {
		return "", err
	}
	accountNum := acct.GetAccountNumber()
	accountSeq := acct.GetSequence()

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sig := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_EIP_712,
			Signature: nil,
		},
		Sequence: accountSeq,
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return "", err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sig = signing.SignatureV2{}

	signerData := xauthsigning.SignerData{
		ChainID:       "inscription_9000-121",
		AccountNumber: accountNum,
		Sequence:      accountSeq,
	}

	sig, err = clitx.SignWithPrivKey(signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		privKey,
		txConfig,
		accountSeq,
	)
	if err != nil {
		return "", err
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return "", err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())

	if err != nil {
		return "", err
	}
	//Broadcast transaction
	txRes, err := s.executor.grpcTxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		})

	fmt.Println("response code: ", txRes.TxResponse.Code) // Should be `0` if the tx is successful
	fmt.Println("response string: ", txRes.TxResponse.String())

	if err != nil {
		return "", err
	}
	return txRes.TxResponse.TxHash, nil
}
func (s *InsTxSender) simulateTx(txBytes []byte) error {
	txClient := s.executor.grpcTxClient

	// We then call the Simulate method on this client.
	grpcRes, err := txClient.Simulate(
		context.Background(),
		&tx.SimulateRequest{
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("Gas info ", grpcRes.GasInfo) // Prints estimated gas used.

	return nil
}

func (s *InsTxSender) getAggregatedSignatureAndValidatorBitset(payload []byte,
	validators []stakingtypes.Validator) ([]byte, *bitset.BitSet, error) {

	var votes []*Vote

	for _, blsPrivKey := range s.cfg.RelayerConfig.BlsPrivateKeys {
		signer, err := NewSigner(common.Hex2Bytes(blsPrivKey))
		if err != nil {
			return nil, nil, err
		}

		var aggregatedPkgs Packages

		for i := 0; i < 10; i++ {
			pkg := GetPackage(1, uint64(i), payload)
			aggregatedPkgs = append(aggregatedPkgs, pkg)
		}
		encBts, _ := rlp.EncodeToBytes(aggregatedPkgs)
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

	valBitset := bitset.From([]uint64{votedAddressSet})

	if err != nil {
		return nil, nil, err
	}
	return aggregatedSignature, valBitset, nil
}
