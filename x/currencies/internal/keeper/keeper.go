// Currencies module keeper stores currency info, issue, withdraw data.
package keeper

import (
	cdcCodec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Module keeper object.
type Keeper struct {
	cdc        *cdcCodec.Codec
	storeKey   sdk.StoreKey
	paramStore params.Subspace
	bankKeeper bank.Keeper
	vmKeeper   common_vm.VMStorage
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Create new currency keeper.
func NewKeeper(cdc *cdcCodec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace, bankKeeper bank.Keeper, vmKeeper common_vm.VMStorage) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramStore: paramSubspace.WithKeyTable(types.ParamKeyTable()),
		bankKeeper: bankKeeper,
		vmKeeper:   vmKeeper,
	}
}
