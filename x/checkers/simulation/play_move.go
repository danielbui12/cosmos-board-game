package simulation

import (
	"math/rand"

	"github.com/alice/checkers/x/checkers/keeper"
	"github.com/alice/checkers/x/checkers/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgPlayMove(
	ak types.AccountKeeper,
	bk types.BankEscrowKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgPlayMove{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the PlayMove simulation

		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "PlayMove simulation not implemented"), nil, nil
	}
}
