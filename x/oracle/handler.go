package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler handles all oracle type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgPostPrice:
			return handleMsgPostPrice(ctx, k, msg)
		case MsgAddOracle:
			return handleMsgAddOracle(ctx, k, msg)
		case MsgSetOracles:
			return handleMsgSetOracles(ctx, k, msg)
		case MsgSetAsset:
			return handleMsgSetAsset(ctx, k, msg)
		case MsgAddAsset:
			return handleMsgAddAsset(ctx, k, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized oracle message type: %T", msg)
		}
	}
}

// price feed questions:
// do proposers need to post the round in the message? If not, how do we determine the round?

// handleMsgPostPrice handles prices posted by oracles.
func handleMsgPostPrice(ctx sdk.Context, k Keeper, msg MsgPostPrice) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := k.ValidatePostPrice(ctx, msg); err != nil {
		return nil, err
	}

	if _, err := k.GetOracle(ctx, msg.AssetCode, msg.From); err != nil {
		return nil, sdkErrors.Wrap(ErrInvalidOracle, msg.From.String())
	}

	if _, err := k.SetPrice(ctx, msg.From, msg.AssetCode, msg.Price, msg.ReceivedAt); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgAddOracle handles AddOracle message.
func handleMsgAddOracle(ctx sdk.Context, k Keeper, msg MsgAddOracle) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, err := k.GetOracle(ctx, msg.Denom, msg.Oracle); err == nil {
		return nil, sdkErrors.Wrap(ErrInvalidOracle, msg.Oracle.String())
	}

	if err := k.AddOracle(ctx, msg.Nominee.String(), msg.Denom, msg.Oracle); err != nil {
		return nil, sdkErrors.Wrap(ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgSetOracles handles SetOracles message.
func handleMsgSetOracles(ctx sdk.Context, k Keeper, msg MsgSetOracles) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, found := k.GetAsset(ctx, msg.Denom); !found {
		return nil, sdkErrors.Wrap(ErrInvalidAsset, msg.Denom)
	}

	if err := k.SetOracles(ctx, msg.Nominee.String(), msg.Denom, msg.Oracles); err != nil {
		return nil, sdkErrors.Wrap(ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgSetAsset handles SetAsset message.
func handleMsgSetAsset(ctx sdk.Context, k Keeper, msg MsgSetAsset) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, found := k.GetAsset(ctx, msg.Denom); !found {
		return nil, sdkErrors.Wrap(ErrInvalidAsset, msg.Denom)
	}

	if err := k.SetAsset(ctx, msg.Nominee.String(), msg.Denom, msg.Asset); err != nil {
		return nil, sdkErrors.Wrap(ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

//  handleMsgAddAsset handles AddUser message.
func handleMsgAddAsset(ctx sdk.Context, k Keeper, msg MsgAddAsset) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, found := k.GetAsset(ctx, msg.Denom); found {
		return nil, sdkErrors.Wrap(ErrExistingAsset, msg.Denom)
	}

	if err := k.AddAsset(ctx, msg.Nominee.String(), msg.Denom, msg.Asset); err != nil {
		return nil, sdkErrors.Wrap(ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
