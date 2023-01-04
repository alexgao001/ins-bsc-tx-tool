package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	RelayerConfig RelayerConfig `json:"relayer_config"`
	InsConfig     InsConfig     `json:"ins_config"`
	BscConfig     BscConfig     `json:"bsc_config"`
}

type RelayerConfig struct {
	RelayerInsPrivateKey string   `json:"relayer_ins_private_key"`
	RelayerBscPrivateKey string   `json:"relayer_bsc_private_key"`
	BlsPrivateKeys       []string `json:"bls_private_keys"`
}

type InsConfig struct {
	GrpcUrl     string `json:"grpc_url"`
	RpcUrl      string `json:"rpc_url"`
	SrcChainId  uint32 `json:"src_chain_id"`
	DestChainId uint32 `json:"dest_chain_id"`
	ToAddress   string `json:"to_address"`
}

type BscConfig struct {
	Provider    string `json:"provider"`
	SrcChainId  uint32 `json:"src_chain_id"`
	DestChainId uint32 `json:"dest_chain_id"`
	ChannelID   int8   `json:"channel_id"`
	GasLimit    uint64 `json:"gas_limit"`
	GasPrice    uint64 `json:"gas_price"`
}

func ParseConfigFromFile(filePath string) *Config {
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var config Config
	if err := json.Unmarshal(bz, &config); err != nil {
		panic(err)
	}

	return &config
}
