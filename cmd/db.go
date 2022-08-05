package cmd

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	SyncPrefix = []byte{0x01}
	PairPrefix = []byte{0x02}
)

func GetSyncKey(height int64) []byte {
	return append(SyncPrefix, sdk.Uint64ToBigEndian(uint64(height))...)
}

func ParseKey(key []byte) int64 {
	return int64(sdk.BigEndianToUint64(key[1:]))
}

func GetPairKey(height int64, pairId uint64) []byte {
	return append(append(PairPrefix, sdk.Uint64ToBigEndian(uint64(height))...), sdk.Uint64ToBigEndian(pairId)...)
}

func ParsePairKey(key []byte) (height int64, pairId uint64) {
	return int64(sdk.BigEndianToUint64(key[1:9])), sdk.BigEndianToUint64(key[9:])
}

func (ctx Context) SyncLog(height int64) {
	ctx.SyncDB.Put(GetSyncKey(height), []byte(time.Now().UTC().String()), nil)
}

func (ctx Context) SyncLogPrint() error {
	iter := ctx.SyncDB.NewIterator(util.BytesPrefix(SyncPrefix), nil)
	for iter.Next() {
		// Use key/value.
		fmt.Println(ParseKey(iter.Key()), string(iter.Value()))
	}
	iter.Release()
	return iter.Error()
}

func (ctx Context) SetPairs(pairs []liquiditytypes.Pair, height int64) {
	for _, pair := range pairs {
		ctx.SyncDB.Put(GetPairKey(height, pair.Id), ctx.Enc.Marshaler.MustMarshal(&pair), nil)
	}
}

func (ctx Context) PairsPrint() error {
	iter := ctx.SyncDB.NewIterator(util.BytesPrefix(PairPrefix), nil)
	for iter.Next() {
		height, pairId := ParsePairKey(iter.Key())
		var pair liquiditytypes.Pair
		ctx.Enc.Marshaler.Unmarshal(iter.Value(), &pair)
		fmt.Println(height, pairId, pair)
	}
	iter.Release()
	return iter.Error()
}
