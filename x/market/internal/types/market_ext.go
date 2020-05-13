package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies_register"
)

// MarketExtended is a Market object extended with currency info from currencies_register module.
// Object is used by order module.
type MarketExtended struct {
	// Market unique ID
	ID dnTypes.ID `json:"id" yaml:"id"`
	// Base asset currency (for ex. btc)
	BaseCurrency currencies_register.CurrencyInfo
	// Quote asset currency (for ex. dfi)
	QuoteCurrency currencies_register.CurrencyInfo
}

// QuantityToDecimal converts quantity to sdk.Dec with currency specifics.
func (m MarketExtended) QuantityToDecimal(quantity sdk.Uint) sdk.Dec {
	return sdk.NewDecFromIntWithPrec(sdk.Int(quantity), int64(m.BaseCurrency.Decimals))
}

// BaseToQuoteQuantity converts base asset price and quantity to quote asset quantity.
// Function normalizes quantity to be used later by OrderBook module, that way quantity for bid and ask
// orders are casted to the same base (base quantity).
func (m MarketExtended) BaseToQuoteQuantity(basePrice sdk.Uint, baseQuantity sdk.Uint) (sdk.Uint, error) {
	pDec := sdk.NewDecFromBigInt(basePrice.BigInt())
	qDec := m.QuantityToDecimal(baseQuantity)

	resDec := pDec.Mul(qDec)
	if resDec.IsZero() {
		return sdk.Uint{}, sdkErrors.Wrap(ErrInvalidQuantity, "quantity is too small")
	}
	resUint := sdk.NewUintFromBigInt(resDec.TruncateInt().BigInt())

	return resUint, nil
}

// BaseDenom return string base asset denom representation.
func (m MarketExtended) BaseDenom() string {
	return string(m.BaseCurrency.Denom)
}

// BaseDenom return string quote asset denom representation.
func (m MarketExtended) QuoteDenom() string {
	return string(m.QuoteCurrency.Denom)
}

// String returns multi-line text object representation.
func (m MarketExtended) String() string {
	b := strings.Builder{}
	b.WriteString("MarketExtended:\n")
	b.WriteString(fmt.Sprintf("  ID: %s\n", m.ID.String()))
	b.WriteString(fmt.Sprintf("  BaseCurrency: %s\n", m.BaseCurrency.String()))
	b.WriteString(fmt.Sprintf("  QuoteCurrency: %s\n", m.QuoteCurrency.String()))

	return b.String()
}

// TableHeaders returns table headers for multi-line text table object representation.
func (m MarketExtended) TableHeaders() []string {
	return []string{
		"M.ID",
		"M.BaseAssetDenom",
		"M.BaseAssetDecimals",
		"M.QuoteAssetDenom",
		"M.QuoteAssetDecimals",
	}
}

// TableHeaders returns table rows for multi-line text table object representation.
func (m MarketExtended) TableValues() []string {
	return []string{
		m.ID.String(),
		m.BaseDenom(),
		strconv.FormatUint(uint64(m.BaseCurrency.Decimals), 10),
		m.QuoteDenom(),
		strconv.FormatUint(uint64(m.QuoteCurrency.Decimals), 10),
	}
}

func NewMarketExtended(market Market, baseCurrency, quoteCurrency currencies_register.CurrencyInfo) MarketExtended {
	return MarketExtended{
		ID:            market.ID,
		BaseCurrency:  baseCurrency,
		QuoteCurrency: quoteCurrency,
	}
}
