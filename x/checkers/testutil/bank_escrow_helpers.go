package testutil

import (
	"context"

	"cosmossdk.io/math"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
)

func (escrow *MockBankEscrowKeeper) ExpectAny(context context.Context) {
	escrow.EXPECT().SendCoinsFromAccountToModule(sdk.UnwrapSDKContext(context), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	escrow.EXPECT().SendCoinsFromModuleToAccount(sdk.UnwrapSDKContext(context), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
}

func coinsOf(amount uint64, denom string) sdk.Coins {
	return sdk.Coins{
		sdk.Coin{
			Denom:  denom,
			Amount: math.NewInt(int64(amount)),
		},
	}
}

func (escrow *MockBankEscrowKeeper) ExpectPay(context context.Context, who string, amount uint64) *gomock.Call {
	return escrow.ExpectPayWithDenom(context, who, amount, sdk.DefaultBondDenom)
}

func (escrow *MockBankEscrowKeeper) ExpectPayWithDenom(context context.Context, who string, amount uint64, denom string) *gomock.Call {
	whoAddr, err := sdk.AccAddressFromBech32(who)
	if err != nil {
		panic(err)
	}
	return escrow.EXPECT().SendCoinsFromAccountToModule(sdk.UnwrapSDKContext(context), whoAddr, types.ModuleName, coinsOf(amount, denom))
}

func (escrow *MockBankEscrowKeeper) ExpectRefund(context context.Context, who string, amount uint64) *gomock.Call {
	return escrow.ExpectRefundWithDenom(context, who, amount, sdk.DefaultBondDenom)
}

func (escrow *MockBankEscrowKeeper) ExpectRefundWithDenom(context context.Context, who string, amount uint64, denom string) *gomock.Call {
	whoAddr, err := sdk.AccAddressFromBech32(who)
	if err != nil {
		panic(err)
	}
	return escrow.EXPECT().SendCoinsFromModuleToAccount(sdk.UnwrapSDKContext(context), types.ModuleName, whoAddr, coinsOf(amount, denom))
}
