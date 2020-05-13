package types

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/olekukonko/tablewriter"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Market object.
// Object is used to store currency references.
type Market struct {
	// Market unique ID
	ID dnTypes.ID `json:"id" yaml:"id"`
	// Base asset denomination (for ex. btc)
	BaseAssetDenom string `json:"base_asset_denom" yaml:"base_asset_denom"`
	// Quote asset denomination (for ex. dfi)
	QuoteAssetDenom string `json:"quote_asset_denom" yaml:"quote_asset_denom"`
}

// Valid check object validity.
func (m Market) Valid() error {
	if err := m.ID.Valid(); err != nil {
		return sdkErrors.Wrap(ErrWrongID, err.Error())
	}
	if err := sdk.ValidateDenom(m.BaseAssetDenom); err != nil {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "BaseAsset")
	}
	if err := sdk.ValidateDenom(m.QuoteAssetDenom); err != nil {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "QuoteAsset")
	}

	return nil
}

// String returns multi-line text object representation.
func (m Market) String() string {
	b := strings.Builder{}
	b.WriteString("Market:\n")
	b.WriteString(fmt.Sprintf("  ID:              %s\n", m.ID.String()))
	b.WriteString(fmt.Sprintf("  BaseAssetDenom:  %s\n", m.BaseAssetDenom))
	b.WriteString(fmt.Sprintf("  QuoteAssetDenom: %s\n", m.QuoteAssetDenom))

	return b.String()
}

// TableHeaders returns table headers for multi-line text table object representation.
func (m Market) TableHeaders() []string {
	return []string{
		"M.ID",
		"M.BaseAssetDenom",
		"M.QuoteAssetDenom",
	}
}

// TableHeaders returns table rows for multi-line text table object representation.
func (m Market) TableValues() []string {
	return []string{
		m.ID.String(),
		m.BaseAssetDenom,
		m.QuoteAssetDenom,
	}
}

// NewMarket create a new market object.
func NewMarket(id dnTypes.ID, baseAsset, quoteAsset string) Market {
	return Market{
		ID:              id,
		BaseAssetDenom:  baseAsset,
		QuoteAssetDenom: quoteAsset,
	}
}

// Market slice type.
type Markets []Market

// Strings returns multi-line text object representation.
func (l Markets) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader(Market{}.TableHeaders())

	for _, m := range l {
		t.Append(m.TableValues())
	}
	t.Render()

	return string(buf.Bytes())
}
