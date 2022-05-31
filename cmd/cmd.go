package cmd

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
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

			db, err := sdk.NewLevelDB("data/blockstore", dir)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			stateDB, err := sdk.NewLevelDB("data/state", dir)
			if err != nil {
				panic(err)
			}
			defer stateDB.Close()

			blockStore := store.NewBlockStore(db)
			stateStore := state.NewStore(stateDB)

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

			//blockList := []string{}
			//swapTxs := []abci.Event{}
			//swapEndBlocks := []abci.Event{}
			//validatorList := []string{}
			for i := startHeight; i < endHeight; i++ {
				if i%10000 == 0 {
					fmt.Println(i)
				}
				//b, err := json.Marshal(blockStore.LoadBlockCommit(i))
				//if err != nil {
				//	panic(err)
				//}
				//blockList = append(blockList, string(b))
				results, err := stateStore.LoadABCIResponses(i)
				if err != nil {
					return err
				}
				for _, i := range results.DeliverTxs {
					for _, j := range i.Events {
						if j.Type == "limit_order" {
							fmt.Println(j.Type, j.String())
						}
					}
				}

				for _, i := range results.EndBlock.Events {
					fmt.Println(i.Type, i.String())
				}
				//results.BeginBlock.Events
			}
			//blockOutput := strings.Join(blockList, "\n")

			//err = ioutil.WriteFile(fmt.Sprintf("blocks-%d-%d.json", startHeight, endHeight), []byte(blockOutput), 0644)
			//if err != nil {
			//	panic(err)
			//}
			//fmt.Println("Done! check the output files on current dir")
			return nil
		},
	}
	return cmd
}
