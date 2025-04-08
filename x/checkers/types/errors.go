package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/checkers module sentinel errors
var (
	ErrInvalidSigner    = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidBlack     = sdkerrors.Register(ModuleName, 1101, "black address is invalid: %s")
	ErrInvalidRed       = sdkerrors.Register(ModuleName, 1102, "red address is invalid: %s")
	ErrGameNotParseable = sdkerrors.Register(ModuleName, 1103, "game cannot be parsed")
	ErrInvalidGameIndex     = sdkerrors.Register(ModuleName, 1104, "game index is invalid")
	ErrInvalidPositionIndex = sdkerrors.Register(ModuleName, 1105, "position index is invalid")
	ErrMoveAbsent           = sdkerrors.Register(ModuleName, 1106, "there is no move")
	ErrGameNotFound         = sdkerrors.Register(ModuleName, 1107, "game by id not found")
	ErrCreatorNotPlayer     = sdkerrors.Register(ModuleName, 1108, "message creator is not a player")
	ErrNotPlayerTurn        = sdkerrors.Register(ModuleName, 1109, "player tried to play out of turn")
	ErrWrongMove            = sdkerrors.Register(ModuleName, 1110, "wrong move")
)
