package types

import "time"

const (
	// ModuleName defines the module name
	ModuleName = "checkers"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_checkers"
)

var (
	ParamsKey = []byte("p_checkers")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	SystemInfoKey = "SystemInfo/value/"
)

const (
	GameCreatedEventType      = "new-game-created" // Indicates what event type to listen to
	GameCreatedEventCreator   = "creator"          // Subsidiary information
	GameCreatedEventGameIndex = "game-index"
	GameCreatedEventBlack     = "black"
	GameCreatedEventRed       = "red"
	GameCreatedEventWager     = "wager"
	GameCreatedEventDenom     = "denom"
)

const (
	MovePlayedEventType      = "move-played"
	MovePlayedEventCreator   = "creator"
	MovePlayedEventGameIndex = "game-index"
	MovePlayedEventCapturedX = "captured-x"
	MovePlayedEventCapturedY = "captured-y"
	MovePlayedEventWinner    = "winner"
	MovePlayedEventBoard     = "board"
)

const (
	MaxTurnDuration = time.Duration(24 * 3_600 * 1000_000_000) // 1 day
	DeadlineLayout  = "2006-01-02 15:04:05.999999999 +0000 UTC"
)

const (
	NoFifoIndex = "-1"
)

const (
	GameForfeitedEventType      = "game-forfeited"
	GameForfeitedEventGameIndex = "game-index"
	GameForfeitedEventWinner    = "winner"
	GameForfeitedEventBoard     = "board"
)

const (
	CreateGameGas = 15000
	PlayMoveGas   = 1000
)
