// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test keeper IssueCurrency method.
func TestCurrenciesKeeper_IssueCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	// ok
	{
		require.False(t, keeper.HasIssue(ctx, defIssueID1))
		require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

		// check account balance changed
		require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).Equal(defAmount))

		// check currency supply increased
		currency, err := keeper.GetCurrency(ctx, defDenom)
		require.NoError(t, err)
		require.True(t, currency.Supply.Equal(defAmount))

		// check currencyInfo supply increased
		curInfo, err := keeper.GetResStdCurrencyInfo(ctx, defDenom)
		require.NoError(t, err)
		require.Equal(t, curInfo.TotalSupply.String(), defAmount.String())
	}

	// fail: existing issueID
	{
		require.Error(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))
	}

	// fail: wrong decimals
	{
		require.Error(t, keeper.IssueCurrency(ctx, defIssueID2, defDenom, defAmount, 2, addr))
	}

	// ok: issue existing currency, increasing supply
	{
		newAmount := defAmount.MulRaw(2)

		require.False(t, keeper.HasIssue(ctx, defIssueID2))
		require.NoError(t, keeper.IssueCurrency(ctx, defIssueID2, defDenom, defAmount, defDecimals, addr))

		// check account balance changed
		require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).Equal(newAmount))

		// check currency supply increased
		currency, err := keeper.GetCurrency(ctx, defDenom)
		require.NoError(t, err)
		require.True(t, currency.Supply.Equal(newAmount))

		// check currencyInfo supply increased
		curInfo, err := keeper.GetResStdCurrencyInfo(ctx, defDenom)
		require.NoError(t, err)
		require.Equal(t, curInfo.TotalSupply.String(), newAmount.String())
	}
}

// Test keeper GetIssue method.
func TestCurrenciesKeeper_GetIssue(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

	// ok
	{
		issue, err := keeper.GetIssue(ctx, defIssueID1)
		require.NoError(t, err)
		require.Equal(t, defDenom, issue.Denom)
		require.True(t, issue.Amount.Equal(defAmount))
		require.Equal(t, addr.String(), issue.Payee.String())
		require.True(t, keeper.HasIssue(ctx, defIssueID1))
	}

	// fail: non-existing
	{
		_, err := keeper.GetIssue(ctx, defIssueID2)
		require.Error(t, err)
		require.False(t, keeper.HasIssue(ctx, defIssueID2))
	}
}

// Test keeper IssueCurrency method: huge amount.
func TestCurrenciesKeeper_IssueHugeAmount(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	amount, ok := sdk.NewIntFromString("1000000000000000000000000000000000000000000000")
	require.True(t, ok)

	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, amount, defDecimals, addr))
	require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).Equal(amount))
}
