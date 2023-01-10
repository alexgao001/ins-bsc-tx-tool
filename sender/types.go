package sender

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type ChainId uint16
type ChannelId uint8

const (
	BSCChainId        ChainId = 2
	prefixLength              = 1
	destChainIDLength         = 2
	channelIDLength           = 1

	DefaultGasPrice = 20000000000 // 20 GWei
)

var (
	crossChainContractAddr = common.HexToAddress("0x0000000000000000000000000000000000002000")
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

type BlsClaim struct {
	SrcChainId  uint32
	DestChainId uint32
	Timestamp   uint64
	Sequence    uint64
	Payload     []byte
}

func (c *BlsClaim) GetSignBytes() [32]byte {
	bts, err := rlp.EncodeToBytes(c)
	if err != nil {
		panic("encode bls claim error")
	}

	btsHash := sdk.Keccak256Hash(bts)
	return btsHash
}
