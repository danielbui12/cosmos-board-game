package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/alice/checkers/testutil/keeper"	
	checkers "github.com/alice/checkers/x/checkers/module"
	"github.com/alice/checkers/x/checkers/testutil"
	"github.com/alice/checkers/x/checkers/keeper"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func setupMsgServerCreateGame(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context) {
	k, ctx := keepertest.CheckersKeeper(t)
	checkers.InitGenesis(ctx, k, *types.DefaultGenesis())
	return keeper.NewMsgServerImpl(k), k, sdk.WrapSDKContext(ctx)
}

func TestCreateGame(t *testing.T) {
	msgServer, _, context := setupMsgServerCreateGame(t)
	createResponse, err := msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: testutil.Alice,
		Black:   testutil.Bob,
		Red:     testutil.Carol,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgCreateGameResponse{
		GameIndex: "1",
	}, *createResponse)
}

func TestCreate1GameHasSaved(t *testing.T) {
	msgSrvr, keeper, context := setupMsgServerCreateGame(t)
	ctx := sdk.UnwrapSDKContext(context)
	msgSrvr.CreateGame(context, &types.MsgCreateGame{
		Creator: testutil.Alice,
		Black:   testutil.Bob,
		Red:     testutil.Carol,
	})
	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId: 2,
	}, systemInfo)
	game1, found1 := keeper.GetStoredGame(ctx, "1")
	require.True(t, found1)
	require.EqualValues(t, types.StoredGame{
		Index: "1",
		Board: "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
		Turn:  "b",
		Black: testutil.Bob,
		Red:   testutil.Carol,
	}, game1)
}