package sender

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"google.golang.org/grpc"
)

func grpcConn(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newRpcClient(addr string) (*rpchttp.HTTP, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}

type InscriptionExecutor struct {
	rpcClient    rpcclient.Client
	queryClient  stakingtypes.QueryClient
	grpcTxClient tx.ServiceClient
	authClient   authtypes.QueryClient
	cdc          *codec.ProtoCodec
}

func NewInscriptionExecutor(rpcUri, grpcUri string) (*InscriptionExecutor, error) {

	rpcClient, err := newRpcClient(rpcUri)
	if err != nil {
		return nil, err
	}
	conn, err := grpcConn(grpcUri)
	if err != nil {
		return nil, err
	}
	queryClient := stakingtypes.NewQueryClient(conn)
	grpcTxClient := tx.NewServiceClient(conn)

	interfaceRegistry := types.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("AccountI", (*authtypes.AccountI)(nil))
	interfaceRegistry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
	)
	interfaceRegistry.RegisterInterface("cosmos.crypto.PubKey", (*cryptotypes.PubKey)(nil))
	interfaceRegistry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ethsecp256k1.PubKey{})

	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &oracletypes.MsgClaim{})

	return &InscriptionExecutor{
		grpcTxClient: grpcTxClient,
		authClient:   authtypes.NewQueryClient(conn),
		rpcClient:    rpcClient,
		queryClient:  queryClient,
		cdc:          codec.NewProtoCodec(interfaceRegistry),
	}, nil
}

func (e *InscriptionExecutor) QueryLatestValidators() ([]stakingtypes.Validator, error) {
	height, err := e.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	result, err := e.queryClient.HistoricalInfo(context.Background(), &stakingtypes.QueryHistoricalInfoRequest{Height: int64(height)})
	if err != nil {
		return nil, err
	}
	hist := result.Hist
	return hist.Valset, nil
}

func (e *InscriptionExecutor) GetLatestBlockHeight() (uint64, error) {
	status, err := e.rpcClient.Status(context.Background())
	if err != nil {
		return 0, err
	}
	return uint64(status.SyncInfo.LatestBlockHeight), nil
}

func (e *InscriptionExecutor) GetNextOracleSequence() (uint64, error) {
	//TODO confirm path to and key to retrive from store
	path := fmt.Sprintf("/store/%s/%s", SequenceStoreName, "key")
	key := BuildChannelSequenceKey(BSCChainId, 0x00)
	response, err := e.rpcClient.ABCIQuery(context.Background(), path, key)
	if err != nil {
		return 0, err
	}
	if response.Response.Value == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(response.Response.Value), nil
}

func (e *InscriptionExecutor) GetAccount(address string) (authtypes.AccountI, error) {
	authRes, err := e.authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}
	var account authtypes.AccountI
	if err := e.cdc.InterfaceRegistry().UnpackAny(authRes.Account, &account); err != nil {
		return nil, err
	}
	return account, nil
}

func (e *InscriptionExecutor) GetAccounts() (authtypes.AccountI, error) {
	accountsRes, err := e.authClient.Accounts(context.Background(), &authtypes.QueryAccountsRequest{})
	if err != nil {
		return nil, err
	}
	for _, a := range accountsRes.GetAccounts() {
		println(hex.EncodeToString(a.Value))
	}

	return nil, nil
}
