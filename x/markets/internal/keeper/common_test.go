// +build unit

package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies_register"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// Mock VM storage implementation.
type VMStorage struct {
	storeKey sdk.StoreKey
}

func NewVMStorage(storeKey sdk.StoreKey) VMStorage {
	return VMStorage{
		storeKey: storeKey,
	}
}

func (storage VMStorage) GetOracleAccessPath(_ string) *vm_grpc.VMAccessPath {
	return &vm_grpc.VMAccessPath{}
}

func (storage VMStorage) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(storage.storeKey)
	store.Set(common_vm.MakePathKey(accessPath), value)
}

func (storage VMStorage) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(storage.storeKey)
	return store.Get(common_vm.MakePathKey(accessPath))
}

func (storage VMStorage) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(storage.storeKey)
	store.Delete(common_vm.MakePathKey(accessPath))
}

func (storage VMStorage) HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(storage.storeKey)
	return store.Has(common_vm.MakePathKey(accessPath))
}

// Module keeper tests input.
type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyParams    *sdk.KVStoreKey
	keyCR        *sdk.KVStoreKey
	keyVMStorage *sdk.KVStoreKey
	tKeyParams   *sdk.TransientStoreKey
	//
	baseBtcDenom    string
	baseBtcDecimals uint8
	baseEthDenom    string
	baseEthDecimals uint8
	quoteDenom      string
	quoteDecimals   uint8
	//
	vmStorage    common_vm.VMStorage
	paramsKeeper params.Keeper
	crKeeper     currencies_register.Keeper
	keeper       Keeper
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:          codec.New(),
		keyParams:    sdk.NewKVStoreKey("key_params"),
		keyCR:        sdk.NewKVStoreKey("key_cr"),
		keyVMStorage: sdk.NewKVStoreKey("key_vm_storage"),
		tKeyParams:   sdk.NewTransientStoreKey("tkey_params"),
		//
		baseBtcDenom:    "btc",
		baseBtcDecimals: 8,
		baseEthDenom:    "eth",
		baseEthDecimals: 18,
		quoteDenom:      "dfi",
		quoteDecimals:   18,
	}

	// register codec
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCR, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyVMStorage, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create target and dependant keepers
	input.vmStorage = NewVMStorage(input.keyVMStorage)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.crKeeper = currencies_register.NewKeeper(input.cdc, input.keyCR, input.vmStorage)
	input.keeper = NewKeeper(input.cdc, input.paramsKeeper.Subspace(types.DefaultParamspace), input.crKeeper)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init params
	input.keeper.SetParams(input.ctx, types.DefaultParams())

	// init currencies
	baseSupply, ok := sdk.NewIntFromString("100000000000000")
	require.True(t, ok)
	quoteSupply, ok := sdk.NewIntFromString("100000000000000000000000000")
	require.True(t, ok)

	ownerAddr := make([]byte, common_vm.VMAddressLength)

	err := input.crKeeper.AddCurrencyInfo(
		input.ctx,
		input.baseBtcDenom,
		input.baseBtcDecimals,
		false,
		ownerAddr,
		baseSupply,
		[]byte("01fe7c965b1c008c5974c7750959fa10189e803225d5057207563553922a09f906"))
	require.NoError(t, err)

	err = input.crKeeper.AddCurrencyInfo(
		input.ctx,
		input.baseEthDenom,
		input.baseEthDecimals,
		false,
		ownerAddr,
		baseSupply,
		[]byte("01f8799f504905a182aff8d5fc102da1d73b8bec199147bb5512af6e99006baeb6"))
	require.NoError(t, err)

	err = input.crKeeper.AddCurrencyInfo(
		input.ctx,
		input.quoteDenom,
		input.quoteDecimals,
		false,
		ownerAddr,
		quoteSupply,
		[]byte("018bfc024222e94fbed60ff0c9c1cf48c5b2809d83c82f513b2c385e21ba8a2d35"))
	require.NoError(t, err)

	return input
}
