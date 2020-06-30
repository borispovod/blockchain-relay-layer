// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/common_vm"
)

// Test MsgIssueCurrency ValidateBasic.
func TestCurrencies_NewCurrencyInfo(t *testing.T) {
	currency := NewCurrency("test", sdk.NewIntFromUint64(100), 8)

	// ok: stdlib
	{
		curInfo, err := NewCurrencyInfo(currency, common_vm.StdLibAddress)
		require.NoError(t, err)
		require.EqualValues(t, currency.Denom, curInfo.Denom)
		require.EqualValues(t, currency.Decimals, curInfo.Decimals)
		require.EqualValues(t, currency.Supply.Uint64(), curInfo.TotalSupply.Uint64())
		require.EqualValues(t, common_vm.StdLibAddress, curInfo.Owner)
		require.False(t, curInfo.IsToken)
	}

	// ok: token
	{
		owner := make([]byte, common_vm.VMAddressLength)

		curInfo, err := NewCurrencyInfo(currency, owner)
		require.NoError(t, err)
		require.EqualValues(t, currency.Denom, curInfo.Denom)
		require.EqualValues(t, currency.Decimals, curInfo.Decimals)
		require.EqualValues(t, currency.Supply.Uint64(), curInfo.TotalSupply.Uint64())
		require.EqualValues(t, owner, curInfo.Owner)
		require.True(t, curInfo.IsToken)
	}

	// fail
	{
		owner := make([]byte, common_vm.VMAddressLength-1)

		_, err := NewCurrencyInfo(currency, owner)
		require.Error(t, err)
	}
}
