package cmd_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/types"

	"github.com/cosmosquad-labs/blockparser/cmd"
)

var (
	orders = []types.Order{
		types.Order{
			Id:        478556,
			PairId:    1,
			MsgHeight: 478556,
			Orderer:   "cre18s7e4ag2stm85jwlvy7fs7hufx8xc0sg3efwuy",
			Direction: types.OrderDirectionSell, // 2
			OfferCoin: sdk.Coin{
				Denom:  "ubcre",
				Amount: sdk.NewInt(100000000),
			},
			RemainingOfferCoin: sdk.Coin{
				Denom:  "ubcre",
				Amount: sdk.NewInt(100000000),
			},
			ReceivedCoin: sdk.Coin{
				Denom:  "ucre",
				Amount: sdk.NewInt(0),
			},
			Price:      sdk.MustNewDecFromStr("1.030000000000000000"),
			Amount:     sdk.NewInt(100000000),
			OpenAmount: sdk.NewInt(100000000),
			BatchId:    399117,
			ExpireAt:   utils.ParseTime("2022-05-16T13:41:22.760931916Z"),
			Status:     types.OrderStatusNotMatched, // 2
		},
		types.Order{
			Id:        478556,
			PairId:    1,
			MsgHeight: 478556,
			Orderer:   "cre18s7e4ag2stm85jwlvy7fs7hufx8xc0sg3efwuy",
			Direction: types.OrderDirectionSell, // 2
			OfferCoin: sdk.Coin{
				Denom:  "ubcre",
				Amount: sdk.NewInt(100000000),
			},
			RemainingOfferCoin: sdk.Coin{
				Denom:  "ubcre",
				Amount: sdk.NewInt(100000000),
			},
			ReceivedCoin: sdk.Coin{
				Denom:  "ucre",
				Amount: sdk.NewInt(0),
			},
			Price:      sdk.MustNewDecFromStr("1.020000000000000000"),
			Amount:     sdk.NewInt(100000000),
			OpenAmount: sdk.NewInt(100000000),
			BatchId:    399117,
			ExpireAt:   utils.ParseTime("2022-05-16T13:41:22.760931916Z"),
			Status:     types.OrderStatusNotMatched, // 2
		},
		types.Order{
			Id:        238050,
			PairId:    1,
			MsgHeight: 478557,
			Orderer:   "cre18s7e4ag2stm85jwlvy7fs7hufx8xc0sg3efwuy",
			Direction: types.OrderDirectionBuy, // 1
			OfferCoin: sdk.Coin{
				Denom:  "ucre",
				Amount: sdk.NewInt(200000000),
			},
			RemainingOfferCoin: sdk.Coin{
				Denom:  "ucre",
				Amount: sdk.NewInt(200000000),
			},
			ReceivedCoin: sdk.Coin{
				Denom:  "ubcre",
				Amount: sdk.NewInt(0),
			},
			Price:      sdk.MustNewDecFromStr("1.000000000000000000"),
			Amount:     sdk.NewInt(200000000),
			OpenAmount: sdk.NewInt(200000000),
			BatchId:    399118,
			ExpireAt:   utils.ParseTime("2022-05-16T13:41:28.708745379Z"),
			Status:     types.OrderStatusNotMatched, // 2
		},
		types.Order{
			Id:        238050,
			PairId:    1,
			MsgHeight: 478557,
			Orderer:   "cre18s7e4ag2stm85jwlvy7fs7hufx8xc0sg3efwuy",
			Direction: types.OrderDirectionBuy, // 1
			OfferCoin: sdk.Coin{
				Denom:  "ucre",
				Amount: sdk.NewInt(300000000),
			},
			RemainingOfferCoin: sdk.Coin{
				Denom:  "ucre",
				Amount: sdk.NewInt(300000000),
			},
			ReceivedCoin: sdk.Coin{
				Denom:  "ubcre",
				Amount: sdk.NewInt(0),
			},
			Price:      sdk.MustNewDecFromStr("1.100000000000000000"),
			Amount:     sdk.NewInt(300000000),
			OpenAmount: sdk.NewInt(300000000),
			BatchId:    399118,
			ExpireAt:   utils.ParseTime("2022-05-16T13:41:28.708745379Z"),
			Status:     types.OrderStatusNotMatched, // 2
		},
	}
)

func TestGetResult(t *testing.T) {
	result := cmd.GetResult(orders)
	fmt.Println(result)
	//orders := []types.Order{
	//	types.Order{
	//		Id:        478556,
	//		PairId:    1,
	//		MsgHeight: 478556,
	//		Orderer:   "cre18s7e4ag2stm85jwlvy7fs7hufx8xc0sg3efwuy",
	//		Direction: types.OrderDirectionSell, // 2
	//		OfferCoin: sdk.Coin{
	//			Denom:  "ubcre",
	//			Amount: sdk.NewInt(105224511248),
	//		},
	//		RemainingOfferCoin: sdk.Coin{
	//			Denom:  "ubcre",
	//			Amount: sdk.NewInt(105224511248),
	//		},
	//		ReceivedCoin: sdk.Coin{
	//			Denom:  "ucre",
	//			Amount: sdk.NewInt(0),
	//		},
	//		Price:      sdk.MustNewDecFromStr("1.030000000000000000"),
	//		Amount:     sdk.NewInt(105224511248),
	//		OpenAmount: sdk.NewInt(105224511248),
	//		BatchId:    399117,
	//		ExpireAt:   utils.ParseTime("2022-05-16T13:41:22.760931916Z"),
	//		Status:     types.OrderStatusNotMatched, // 2
	//	},
	//	types.Order{
	//		Id:        238050,
	//		PairId:    1,
	//		MsgHeight: 478557,
	//		Orderer:   "cre18s7e4ag2stm85jwlvy7fs7hufx8xc0sg3efwuy",
	//		Direction: types.OrderDirectionBuy, // 1
	//		OfferCoin: sdk.Coin{
	//			Denom:  "ucre",
	//			Amount: sdk.NewInt(62832999),
	//		},
	//		RemainingOfferCoin: sdk.Coin{
	//			Denom:  "ucre",
	//			Amount: sdk.NewInt(62832999),
	//		},
	//		ReceivedCoin: sdk.Coin{
	//			Denom:  "ubcre",
	//			Amount: sdk.NewInt(0),
	//		},
	//		Price:      sdk.MustNewDecFromStr("1.000000000000000000"),
	//		Amount:     sdk.NewInt(62832999),
	//		OpenAmount: sdk.NewInt(62832999),
	//		BatchId:    399118,
	//		ExpireAt:   utils.ParseTime("2022-05-16T13:41:28.708745379Z"),
	//		Status:     types.OrderStatusNotMatched, // 2
	//	},
	//}
	fmt.Println(orders)
}
