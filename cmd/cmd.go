package cmd

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
)

func NewBlockParserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "blockparser [chain-dir] [start-height] [end-height] [search-string]",
		Args: cobra.ExactArgs(4),
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
			searchStr := args[3]

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

			blockStore := store.NewBlockStore(blockDB)
			stateStore := state.NewStore(stateDB)
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
			for i := startHeight; i < endHeight; i++ {
				//if i%10000 == 0 {
				//	fmt.Println(i)
				//}
				results, err := stateStore.LoadABCIResponses(i)
				if err != nil {
					return err
				}
				for _, tx := range results.DeliverTxs {
					txStr := tx.String()
					if strings.Contains(txStr, searchStr) {
						log, err := sdk.ParseABCILogs(tx.Log)
						logStr := ""
						if err != nil {
							logStr = txStr
						}
						logStr = log.String()
						fmt.Println(i, "[txs]", logStr)
					}
				}

				for _, event := range results.EndBlock.Events {
					if strings.Contains(event.String(), searchStr) {
						fmt.Println(i, "[beginblock]", event.String())
					}

				}
				for _, event := range results.EndBlock.Events {
					if strings.Contains(event.String(), searchStr) {
						fmt.Println(i, "[endblock]", event.String())
					}
				}
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
