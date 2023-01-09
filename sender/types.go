package sender

import "github.com/ethereum/go-ethereum/common"

type ChainId uint16
type ChannelId uint8

const (
	BSCChainId        ChainId = 2
	prefixLength              = 1
	destChainIDLength         = 2
	channelIDLength           = 1

	DefaultGasPrice   = 20000000000 // 20 GWei
	SequenceStoreName = "crosschain"
)

var (
	tendermintLightClientContractAddr = common.HexToAddress("0x0000000000000000000000000000000000001003")
	crossChainContractAddr            = common.HexToAddress("0x0000000000000000000000000000000000002000")
)

var (
	PrefixForReceiveSequenceKey = []byte{0xf1}
)

type MsgClaim struct {
	FromAddress    string
	SrcChainId     uint32
	DestChainId    uint32
	Sequence       uint64
	TimeStamp      uint64
	Payload        []byte
	VoteAddressSet []uint64
	AggSignature   []byte
}

type Packages []Package

type Package struct {
	ChannelId uint8
	Sequence  uint64
	Payload   []byte
}
