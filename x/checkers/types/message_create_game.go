package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateGame{}

func NewMsgCreateGame(creator string, black string, red string, wager uint64, denom string) *MsgCreateGame {
	return &MsgCreateGame{
		Creator: creator,
		Black:   black,
		Red:     red,
		Wager:   wager,
		Denom:   denom,
	}
}

func (msg *MsgCreateGame) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Black)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid black address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Red)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid red address (%s)", err)
	}

	if msg.Wager <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "wager must be greater than 0")
	}

	if msg.Denom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "denom must be non-empty")
	}

	return nil
}
