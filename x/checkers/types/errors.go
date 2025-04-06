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
)
