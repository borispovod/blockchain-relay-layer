package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

func (k Keeper) CreateCurrency(ctx sdk.Context, denom string, params types.CurrencyParams) error {
	if k.HasCurrency(ctx, denom) {
		return sdkErrors.Wrapf(types.ErrWrongDenom, "currency %q: exists", denom)
	}

	// build currency objects
	currency := types.NewCurrency(denom, sdk.ZeroInt(), params.Decimals)
	_, err := types.NewResCurrencyInfo(currency, common_vm.StdLibAddress)
	if err != nil {
		return sdkErrors.Wrapf(types.ErrWrongParams, "currency %q: %v", denom, err)
	}

	// store VM path objects
	k.storeCurrencyBalancePath(ctx, denom, params.BalancePath())
	k.storeCurrencyInfoPath(ctx, denom, params.InfoPath())

	// store currency objects
	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)
	k.updateCurrenciesParams(ctx, denom, params)

	return nil
}

// HasCurrency checks that currency exists.
func (k Keeper) HasCurrency(ctx sdk.Context, denom string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetCurrencyKey(denom))
}

// GetCurrency returns currency.
func (k Keeper) GetCurrency(ctx sdk.Context, denom string) (types.Currency, error) {
	if !k.HasCurrency(ctx, denom) {
		return types.Currency{}, sdkErrors.Wrapf(types.ErrWrongDenom, "currency with %q denom: not found", denom)
	}

	return k.getCurrency(ctx, denom), nil
}

// getCurrency returns currency from the storage
func (k Keeper) getCurrency(ctx sdk.Context, denom string) types.Currency {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCurrencyKey(denom))

	currency := types.Currency{}
	k.cdc.MustUnmarshalBinaryBare(bz, &currency)

	return currency
}

// storeCurrency sets currency to the storage.
func (k Keeper) storeCurrency(ctx sdk.Context, currency types.Currency) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetCurrencyKey(currency.Denom), k.cdc.MustMarshalBinaryBare(currency))
}

// increaseSupply Increases currency supply.
func (k Keeper) increaseSupply(ctx sdk.Context, denom string, amount sdk.Int) {
	currency := k.getCurrency(ctx, denom)
	currency.Supply = currency.Supply.Add(amount)

	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)
}

// reduceSupply reduces currency supply and stores withdraw info.
func (k Keeper) reduceSupply(ctx sdk.Context, denom string, amount sdk.Int, spender sdk.AccAddress, recipient, chainID string) {
	currency := k.getCurrency(ctx, denom)
	currency.Supply = currency.Supply.Sub(amount)

	newId := k.getNextWithdrawID(ctx)
	withdraw := types.NewWithdraw(newId, denom, amount, spender, recipient, chainID, ctx.BlockHeader().Time.Unix(), ctx.TxBytes())

	k.storeWithdraw(ctx, withdraw)
	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)
	k.setLastWithdrawID(ctx, newId)
}
