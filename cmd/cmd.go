package cmd

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb/opt"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	txkv "github.com/tendermint/tendermint/state/txindex/kv"
	"github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"

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

			// Init app
			encCfg := app.MakeEncodingConfig()
			app := app.NewApp(log.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, encCfg, app.EmptyAppOptions{})
			if err := app.LoadHeight(startHeight); err != nil {
				panic(err)
			}

			// Set tx index
			store, err := tmdb.NewGoLevelDBWithOpts("data/tx_index", dir, &opt.Options{
				ErrorIfMissing: true,
				ReadOnly:       true,
			})
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer store.Close()
			txi := txkv.NewTxIndex(store)
			txDecoder := encCfg.TxConfig.TxDecoder()

			//ctx := app.BaseApp.NewContext(true, tmproto.Header{})
			//pairs := app.LiquidityKeeper.GetAllPairs(ctx)
			startTime := time.Now()
			orderCount := 0
			orderTxCount := 0

			// Height => PairId => Order
			orderMap := map[int64]map[uint64][]liquiditytypes.Order{}
			// TODO: convert to mmOrder and indexing by mm address
			mmOrderMap := map[int64]map[uint64][]liquiditytypes.MsgLimitOrder{}
			// TODO: indexing by address
			//mmOrderCancelMap := map[int64]map[uint64]liquiditytypes.MsgLimitOrder{}

			pairsReq := liquiditytypes.QueryPairsRequest{
				Pagination: &query.PageRequest{
					Limit: 1000000,
				},
			}
			pairsRewData, err := pairsReq.Marshal()
			if err != nil {
				return err
			}

			// iterate from startHeight to endHeight
			for i := startHeight; i < endHeight; i++ {
				if i%10000 == 0 {
					fmt.Println(i, time.Now().Sub(startTime), orderCount, orderTxCount)
				}
				orderMap[i] = map[uint64][]liquiditytypes.Order{}
				mmOrderMap[i] = map[uint64][]liquiditytypes.MsgLimitOrder{}

				// Query paris
				pairsRes := app.Query(abci.RequestQuery{
					Path:   "/crescent.liquidity.v1beta1.Query/Pairs",
					Data:   pairsRewData,
					Height: i,
					Prove:  false,
				})
				var pairsLive liquiditytypes.QueryPairsResponse
				pairsLive.Unmarshal(pairsRes.Value)
				for _, pair := range pairsLive.Pairs {
					// TODO: check pruning

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

					// Query Orders
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
						orderMap[i][pair.Id] = append(orderMap[i][pair.Id], order)
					}
				}

				// get block result
				block := blockStore.LoadBlock(i)

				// iterate and parse txs of this block, ordered by txResult.Index 0 -> n
				for _, tx := range block.Txs {
					txResult, err := txi.Get(tx.Hash())
					if err != nil {
						return fmt.Errorf("get tx index: %w", err)
					}

					// pass if not succeeded tx
					if txResult.Result.Code != 0 {
						continue
					}

					sdkTx, err := txDecoder(txResult.Tx)
					if err != nil {
						return fmt.Errorf("decode tx: %w", err)
					}

					// indexing only targeted msg types
					for _, msg := range sdkTx.GetMsgs() {
						switch msg := msg.(type) {
						// TODO: filter only MM order type MMOrder, MMOrderCancel
						case *liquiditytypes.MsgLimitOrder:
							orderTxCount++
							mmOrderMap[i][msg.PairId] = append(mmOrderMap[i][msg.PairId], *msg)
							//fmt.Println(i, msg.Orderer, msg.Price, txResult.Result.Code, hex.EncodeToString(tx.Hash()))
						}
					}
				}
			}
			return nil
		},
	}
	return cmd
}
