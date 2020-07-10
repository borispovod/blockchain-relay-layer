// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestOracleMsg_AddAsset(t *testing.T) {
	t.Parallel()

	nominee := sdk.AccAddress([]byte("someName"))
	assetCode := "btc_dfi"
	oracles := Oracles([]Oracle{NewOracle(sdk.AccAddress([]byte("someName")))})
	asset := NewAsset(assetCode, oracles, true)
	denom := "btc"

	t.Run("MsgInterface", func(t *testing.T) {
		target := NewMsgAddAsset(nominee, denom, asset)
		require.Equal(t, "add_asset", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{nominee}, target.GetSigners())
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		// ok
		{
			msg := NewMsgAddAsset(nominee, denom, asset)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid denom
		{
			msg := NewMsgAddAsset(nominee, "", asset)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid nominee
		{
			msg := NewMsgAddAsset(sdk.AccAddress{}, denom, asset)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid asset
		{
			msg := NewMsgAddAsset(nominee, denom, Asset{})
			require.Error(t, msg.ValidateBasic())
		}
	})
}
