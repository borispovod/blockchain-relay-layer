package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

var (
	MsgDeployModuleType  = "deploy_module"
	MsgExecuteScriptType = "execute_script"

	_ sdk.Msg = MsgDeployModule{}
	_ sdk.Msg = MsgExecuteScript{}
)

// Message to deploy contract.
type MsgDeployModule struct {
	Signer sdk.AccAddress `json:"signer"`
	Module Contract       `json:"module"`
}

func NewMsgDeployModule(signer sdk.AccAddress, module Contract) MsgDeployModule {
	return MsgDeployModule{
		Signer: signer,
		Module: module,
	}
}

func (MsgDeployModule) Route() string {
	return RouterKey
}

func (MsgDeployModule) Type() string {
	return MsgDeployModuleType
}

func (msg MsgDeployModule) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkErrors.Wrapf(sdkErrors.ErrInvalidAddress, "empty deployer address")
	}

	if len(msg.Module) == 0 {
		return ErrEmptyContract
	}

	return nil
}

func (msg MsgDeployModule) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgDeployModule) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// Arguments to execute script.
type ScriptArg struct {
	Type  vm_grpc.VMTypeTag `json:"type"`
	Value []byte            `json:"value"`
}

// New ScriptArg from arguments.
func NewScriptArg(typeTag vm_grpc.VMTypeTag, value []byte) ScriptArg {
	return ScriptArg{
		Type:  typeTag,
		Value: value,
	}
}

// Message for contract script (execution).
type MsgExecuteScript struct {
	Signer sdk.AccAddress `json:"signer"`
	Script Contract       `json:"script"`
	Args   []ScriptArg    `json:"args"`
}

func NewMsgExecuteScript(signer sdk.AccAddress, script Contract, args []ScriptArg) MsgExecuteScript {
	return MsgExecuteScript{
		Signer: signer,
		Script: script,
		Args:   args,
	}
}

func (MsgExecuteScript) Route() string {
	return RouterKey
}

func (MsgExecuteScript) Type() string {
	return MsgExecuteScriptType
}

func (msg MsgExecuteScript) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty deployer address")
	}

	if len(msg.Script) == 0 {
		return ErrEmptyContract
	}

	for _, arg := range msg.Args {
		if _, err := VMTypeTagToString(arg.Type); err != nil {
			return sdkErrors.Wrap(ErrWrongArgTypeTag, err.Error())
		}
		if len(arg.Value) == 0 {
			return ErrWrongArgValue
		}
	}

	return nil
}

func (msg MsgExecuteScript) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgExecuteScript) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
