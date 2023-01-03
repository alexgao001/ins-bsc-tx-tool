package sender

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	ethHd "github.com/evmos/ethermint/crypto/hd"
)

func HexToEthSecp256k1PrivKey(hexString string) (*ethsecp256k1.PrivKey, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}
	return ethHd.EthSecp256k1.Generate()(bz).(*ethsecp256k1.PrivKey), nil
}

func BuildChannelSequenceKey(destChainId ChainId, chanelId ChannelId) []byte {
	key := make([]byte, prefixLength+destChainIDLength+channelIDLength)
	copy(key[:prefixLength], prefixForSequenceKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+destChainIDLength], uint16(destChainId))
	copy(key[prefixLength+destChainIDLength:], []byte{byte(chanelId)})
	return key
}

func GetPackage(channelId uint8, seq uint64, payload []byte) Package {
	return Package{
		ChannelId: channelId,
		Sequence:  seq,
		Payload:   payload,
	}
}
