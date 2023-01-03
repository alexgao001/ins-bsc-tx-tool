package main

import (
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"ins-bsc-tx-tool/config"
	"ins-bsc-tx-tool/sender"
)

const (
	flagTxOnChain = "tx-chain"
	BSC           = "bsc"
	Inscription   = "ins"
)

func initFlags() {
	flag.String(flagTxOnChain, "", "the chain send tx to, bsc")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}
}

func main() {
	initFlags()

	var cfg *config.Config
	txOnChain := viper.GetString(flagTxOnChain)
	cfg = config.ParseConfigFromFile("config/config.json")
	if txOnChain == Inscription {
		err := initInsTxSender(cfg)
		if err != nil {
			panic(err)
		}
	} else if txOnChain == BSC {
		err := initBscTxSender(cfg)
		if err != nil {
			panic(err)
		}
	} else {
		panic("Please specify the correct chain, 'bsc/inscription'")
	}
}

func initInsTxSender(cfg *config.Config) error {

	executor, err := sender.NewInscriptionExecutor(cfg.InsConfig.RpcUrl, cfg.InsConfig.GrpcUrl)
	if err != nil {
		panic(err)
	}

	txSender := sender.NewInsTxSender(executor, cfg)
	txHash, err := txSender.Send()
	if err != nil {
		return err
	}
	fmt.Printf("tx hash is %s", txHash)
	return nil
}

func initBscTxSender(cfg *config.Config) error {

	bscE, err := sender.NewBSCExecutor(cfg)
	if err != nil {
		return err
	}

	insE, err := sender.NewInscriptionExecutor(cfg.InsConfig.RpcUrl, cfg.InsConfig.GrpcUrl)
	if err != nil {
		panic(err)
	}

	s := sender.NewBscTxSender(bscE, insE, cfg)
	txHash, err := s.Send()
	if err != nil {
		return err
	}
	fmt.Printf("tx hash is %s", txHash)
	return nil
}
