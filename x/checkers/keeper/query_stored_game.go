package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/alice/checkers/x/checkers/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) StoredGameAll(ctx context.Context, req *types.QueryAllStoredGameRequest) (*types.QueryAllStoredGameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var storedGames []types.StoredGame

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	storedGameStore := prefix.NewStore(store, types.KeyPrefix(types.StoredGameKeyPrefix))

	// Initialize pagination if not provided
	if req.Pagination == nil {
		req.Pagination = &query.PageRequest{
			Limit: query.DefaultLimit,
		}
	}

	pageRes, err := query.Paginate(storedGameStore, req.Pagination, func(key []byte, value []byte) error {
		var storedGame types.StoredGame
		if err := k.cdc.Unmarshal(value, &storedGame); err != nil {
			return err
		}

		storedGames = append(storedGames, storedGame)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllStoredGameResponse{StoredGame: storedGames, Pagination: pageRes}, nil
}

func (k Keeper) StoredGame(ctx context.Context, req *types.QueryGetStoredGameRequest) (*types.QueryGetStoredGameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetStoredGame(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetStoredGameResponse{StoredGame: val}, nil
}
