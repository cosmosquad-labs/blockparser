package cmd

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/store"

	"github.com/crescent-network/crescent/app"
)

func NewBlockParserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "blockparser [chain-dir] [start-height] [end-height]",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			startHeight, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("parse start-Height: %w", err)
			}

			endHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("parse end-Height: %w", err)
			}

			blockDB, err := sdk.NewLevelDB("data/blockstore", dir)
			if err != nil {
				panic(err)
			}
			defer blockDB.Close()

			stateDB, err := sdk.NewLevelDB("data/state", dir)
			if err != nil {
				panic(err)
			}
			defer stateDB.Close()

			txDB, err := sdk.NewLevelDB("data/tx_index", dir)
			if err != nil {
				panic(err)
			}
			defer txDB.Close()

			db, err := sdk.NewLevelDB("data/application", dir)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			blockStore := store.NewBlockStore(blockDB)
			//stateStore := state.NewStore(stateDB)
			//txStore := kv.NewTxIndex(txDB)

			fmt.Println("Loaded : ", dir+"/data/")
			fmt.Println("Input Start Height :", startHeight)
			fmt.Println("Input End Height :", endHeight)
			fmt.Println("Latest Height :", blockStore.Height())

			// checking start height
			block := blockStore.LoadBlock(startHeight)
			if block == nil {
				fmt.Println(startHeight, "is not available on this data")
				for i := 0; i < 1000000000000; i++ {
					block := blockStore.LoadBlock(int64(i))
					if block != nil {
						fmt.Println("available starting Height : ", i)
						break
					}
				}
				return nil
			}

			// checking end height
			if endHeight > blockStore.Height() {
				fmt.Println(endHeight, "is not available, Latest Height : ", blockStore.Height())
				return nil
			}

			encCfg := app.MakeEncodingConfig()
			app := app.NewApp(log.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, encCfg, app.EmptyAppOptions{})
			if err := app.LoadHeight(startHeight); err != nil {
				panic(err)
			}
			ctx := app.BaseApp.NewContext(true, tmproto.Header{})
			pairs := app.LiquidityKeeper.GetAllPairs(ctx)
			startTime := time.Now()
			orderCount := 0
			// Height => PairId => Order
			orderMap := map[int64]map[uint64]liquiditytypes.Order{}
			for i := startHeight; i < endHeight; i++ {
				if i%10000 == 0 {
					fmt.Println(i, time.Now().Sub(startTime), orderCount)
				}
				orderMap[i] = map[uint64]liquiditytypes.Order{}

				for _, pair := range pairs {
					a := liquiditytypes.QueryOrdersRequest{
						PairId: pair.Id,
						Pagination: &query.PageRequest{
							Limit: 1000000,
						},
					}
					data, err := a.Marshal()
					if err != nil {
						return err
					}
					res := app.Query(abci.RequestQuery{
						Path:   "/crescent.liquidity.v1beta1.Query/Orders",
						Data:   data,
						Height: i,
						Prove:  false,
					})
					var orders liquiditytypes.QueryOrdersResponse
					orders.Unmarshal(res.Value)
					for _, order := range orders.Orders {
						orderCount++
						orderMap[i][pair.Id] = order
						//fmt.Println(order.Id, order.Orderer, order.Amount, order.Price, i)
					}
				}
			}
			return nil
		},
	}
	return cmd
}
