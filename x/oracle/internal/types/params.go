package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/params"
)

var (
	KeyAssets    = []byte("oracleassets")
	KeyNominees  = []byte("oraclenominees")
	KeyPostPrice = []byte("oraclepostprice")
)

// Params defines keeper params.
type Params struct {
	// All assets
	Assets Assets `json:"assets" yaml:"assets"`
	// Nominees addresses
	Nominees []string `json:"nominees" yaml:"nominees"`
	// PostPrice params
	PostPrice PostPriceParams `json:"post_price" yaml:"post_price"`
}

// Implements subspace.ParamSet interface.
func (p Params) ParamSetPairs() params.ParamSetPairs {
	nilPairValidatorFunc := func(value interface{}) error {
		return nil
	}

	return params.ParamSetPairs{
		{Key: KeyAssets, Value: &p.Assets, ValidatorFn: nilPairValidatorFunc},
		{Key: KeyNominees, Value: &p.Nominees, ValidatorFn: nilPairValidatorFunc},
		{Key: KeyPostPrice, Value: &p.PostPrice, ValidatorFn: nilPairValidatorFunc},
	}
}

// Validate ensure that params have valid values.
func (p Params) Validate() error {
	for _, asset := range p.Assets {
		if err := asset.AssetCode.Validate(); err != nil {
			return fmt.Errorf("invalid asset %q: %w", asset.String(), err)
		}
	}

	for i, nominee := range p.Nominees {
		if nominee == "" {
			return fmt.Errorf("invalid nominee [%d]: empty", i)
		}
	}

	return nil
}

func (p Params) String() string {
	out := strings.Builder{}

	out.WriteString("Params:\n")
	for i, a := range p.Assets {
		out.WriteString(fmt.Sprintf("Asset [%d]: %s\n", i, a.String()))
	}
	for i, n := range p.Nominees {
		out.WriteString(fmt.Sprintf("Nominee [%d]: %s\n", i, n))
	}
	out.WriteString(p.PostPrice.String())

	return strings.TrimSpace(out.String())
}

// NewParams creates a new AssetParams object.
func NewParams(assets []Asset, nominees []string, postPrice PostPriceParams) Params {
	return Params{
		Assets:    assets,
		Nominees:  nominees,
		PostPrice: postPrice,
	}
}

// DefaultParams default params for oracle.
func DefaultParams() Params {
	return NewParams(Assets{}, []string{}, PostPriceParams{
		ReceivedAtDiffInS: 60 * 60,
	})
}

// ParamKeyTable Key declaration for parameters.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// PostPriceParams Posting rawPrices from oracles configuration params.
type PostPriceParams struct {
	// Allowed timestamp difference between current block time and oracle's receivedAt (0 - disabled) [sec]
	ReceivedAtDiffInS uint32 `json:"received_at_diff_in_s" yaml:"received_at_diff_in_s"`
}

func (p PostPriceParams) String() string {
	return fmt.Sprintf("PostPrice params:\n"+
		"  ReceivedAtDiffInS: %d",
		p.ReceivedAtDiffInS,
	)
}
