package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/alice/checkers/testutil/keeper"
	"github.com/alice/checkers/x/checkers/keeper"
	checkers "github.com/alice/checkers/x/checkers/module"
	"github.com/alice/checkers/x/checkers/testutil"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func setupMsgServerWithOneGameForPlayMove(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context) {
	server, k, context, _, escrow := setupMsgServerWithOneGameForPlayMoveWithMock(t)
	escrow.ExpectAny(context)
	return server, k, context
}

func setupMsgServerWithOneGameForPlayMoveWithMock(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context,
	*gomock.Controller, *testutil.MockBankEscrowKeeper) {
	ctrl := gomock.NewController(t)
	bankMock := testutil.NewMockBankEscrowKeeper(ctrl)
	k, ctx := keepertest.CheckersKeeperWithMocks(t, bankMock)
	checkers.InitGenesis(ctx, k, *types.DefaultGenesis())
	server := keeper.NewMsgServerImpl(k)
	context := sdk.WrapSDKContext(ctx)
	server.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Black:   bob,
		Red:     carol,
		Wager:   45,
		Denom:   "stake",
	})
	// Clear events after game creation
	sdk.UnwrapSDKContext(context).EventManager().EmitEvents([]sdk.Event{})
	return server, k, context, ctrl, bankMock
}

func TestPlayMove(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: -1,
		CapturedY: -1,
		Winner:    "*",
	}, *playMoveResponse)
}

func TestPlayMoveGameNotFound(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "2",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, "2: game by id not found", err.Error())
}

func TestPlayMoveSameBlackRed(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Black:   bob,
		Red:     bob,
		Wager:   45,
		Denom:   "stake",
	})
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: -1,
		CapturedY: -1,
		Winner:    "*",
	}, *playMoveResponse)
}

func TestPlayMoveSavedGame(t *testing.T) {
	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "1",
		FifoTailIndex: "1",
	}, systemInfo)
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "1",
		Board:       "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
		Turn:        "r",
		Black:       bob,
		Red:         carol,
		Winner:      "*",
		Deadline:    types.FormatDeadline(ctx.BlockTime().Add(types.MaxTurnDuration)),
		MoveCount:   1,
		BeforeIndex: types.NoFifoIndex,
		AfterIndex:  types.NoFifoIndex,
		Wager:       45,
		Denom:       "stake",
	}, game1)
}

func TestPlayMoveCalledBank(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMoveWithMock(t)
	defer ctrl.Finish()
	escrow.ExpectPay(context, bob, 45).Times(1)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
}

func TestPlayMoveNotPlayer(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   alice,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, alice+": message creator is not a player", err.Error())
}

func TestPlayMoveCannotParseGame(t *testing.T) {
	msgServer, k, context := setupMsgServerWithOneGameForPlayMove(t)
	ctx := sdk.UnwrapSDKContext(context)
	storedGame, _ := k.GetStoredGame(ctx, "1")
	storedGame.Board = "not a board"
	k.SetStoredGame(ctx, storedGame)
	defer func() {
		r := recover()
		require.NotNil(t, r, "The code did not panic")
		require.Equal(t, r, "game cannot be parsed: invalid board string: not a board")
	}()
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
}

func TestPlayMoveWrongOutOfTurn(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, "{red}: player tried to play out of turn", err.Error())
}

func TestPlayMoveWrongPieceAtDestination(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     0,
		ToX:       0,
		ToY:       1,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, "Already piece at destination position: {0 1}: wrong move", err.Error())
}

func TestPlayMove2(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: -1,
		CapturedY: -1,
		Winner:    "*",
	}, *playMoveResponse)
}

func TestPlayMove2SavedGame(t *testing.T) {
	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "1",
		FifoTailIndex: "1",
	}, systemInfo)
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "1",
		Board:       "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*",
		Turn:        "b",
		Black:       bob,
		Red:         carol,
		Winner:      "*",
		Deadline:    types.FormatDeadline(ctx.BlockTime().Add(types.MaxTurnDuration)),
		MoveCount:   2,
		BeforeIndex: types.NoFifoIndex,
		AfterIndex:  types.NoFifoIndex,
		Wager:       45,
		Denom:       "stake",
	}, game1)
}

func TestPlayMove3(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     2,
		FromY:     3,
		ToX:       0,
		ToY:       5,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: 1,
		CapturedY: 4,
		Winner:    "*",
	}, *playMoveResponse)
}

func TestPlayMove3SavedGame(t *testing.T) {
	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     2,
		FromY:     3,
		ToX:       0,
		ToY:       5,
	})
	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "1",
		FifoTailIndex: "1",
	}, systemInfo)
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "1",
		Board:       "*b*b*b*b|b*b*b*b*|***b*b*b|********|********|b*r*r*r*|*r*r*r*r|r*r*r*r*",
		Turn:        "r",
		Black:       bob,
		Red:         carol,
		Winner:      "*",
		Deadline:    types.FormatDeadline(ctx.BlockTime().Add(types.MaxTurnDuration)),
		MoveCount:   3,
		BeforeIndex: types.NoFifoIndex,
		AfterIndex:  types.NoFifoIndex,
		Wager:       45,
		Denom:       "stake",
	}, game1)
}

func TestPlayMoveEmitted(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2) // We expect both new-game-created and move-played events

	// Find the move-played event
	var moveEvent sdk.StringEvent
	for _, event := range events {
		if event.Type == "move-played" {
			moveEvent = event
			break
		}
	}
	require.NotEmpty(t, moveEvent, "move-played event not found")

	require.EqualValues(t, sdk.StringEvent{
		Type: "move-played",
		Attributes: []sdk.Attribute{
			{Key: "creator", Value: bob},
			{Key: "game-index", Value: "1"},
			{Key: "captured-x", Value: "-1"},
			{Key: "captured-y", Value: "-1"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, moveEvent)
}
