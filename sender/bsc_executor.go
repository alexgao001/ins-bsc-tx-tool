package sender

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"ins-bsc-tx-tool/config"
	"ins-bsc-tx-tool/crosschain"
	"math/big"
	"time"
)

func getPrivateKey(cfg *config.Config) (*ecdsa.PrivateKey, error) {
	privKey, err := crypto.HexToECDSA(cfg.RelayerConfig.RelayerBscPrivateKey)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

type BSCExecutor struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	cfg        *config.Config
	gasPrice   *big.Int
	txSender   common.Address
}

func NewBSCExecutor(cfg *config.Config) (*BSCExecutor, error) {
	privKey, err := getPrivateKey(cfg)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("get public key error")
	}
	txSender := crypto.PubkeyToAddress(*publicKeyECDSA)
	client, err := ethclient.Dial(cfg.BscConfig.Provider)
	if err != nil {
		return nil, err
	}
	var initGasPrice *big.Int
	if cfg.BscConfig.GasPrice == 0 {
		initGasPrice = big.NewInt(DefaultGasPrice)
	} else {
		initGasPrice = big.NewInt(int64(cfg.BscConfig.GasPrice))
	}
	return &BSCExecutor{
		client:     client,
		privateKey: privKey,
		cfg:        cfg,
		gasPrice:   initGasPrice,
		txSender:   txSender,
	}, nil
}

func (e *BSCExecutor) GetLatestBlockHeight() (uint64, error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	block, err := e.client.BlockByNumber(ctxWithTimeout, nil)
	if err != nil {
		return 0, err
	}
	return block.Number().Uint64(), nil
}

func (e *BSCExecutor) getTransactor(nonce uint64) (*bind.TransactOpts, error) {
	txOpts := bind.NewKeyedTransactor(e.privateKey)
	txOpts.Nonce = big.NewInt(int64(nonce))
	txOpts.Value = big.NewInt(0)
	txOpts.GasLimit = e.cfg.BscConfig.GasLimit
	txOpts.GasPrice = e.gasPrice
	return txOpts, nil
}

func (e *BSCExecutor) getCallOpts() (*bind.CallOpts, error) {
	callOpts := &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	return callOpts, nil
}

func (e *BSCExecutor) GetNextSequence() (uint64, error) {
	crossChainInstance, err := crosschain.NewCrosschain(crossChainContractAddr, e.client)
	if err != nil {
		return 0, err
	}
	callOpts, err := e.getCallOpts()
	if err != nil {
		return 0, err
	}
	return crossChainInstance.ChannelReceiveSequenceMap(callOpts, uint8(e.cfg.BscConfig.ChannelID))
}
