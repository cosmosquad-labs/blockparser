package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	crecmd "github.com/crescent-network/crescent/v2/cmd/crescentd/cmd"

	"github.com/cosmosquad-labs/blockparser/cmd"
)

type CmdTestSuite struct {
	suite.Suite
}

func TestCmdTestSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (suite *CmdTestSuite) SetupTest() {
	crecmd.GetConfig()
}

func (suite *CmdTestSuite) TestMain2() {
	for _, tc := range []struct {
		dir         string
		startHeight int64
		endHeight   int64
	}{
		{
			dir:         "/Users/dongsamb/.crescent",
			startHeight: 478559,
			endHeight:   478566,
		},
	} {
		suite.Run(tc.dir, func() {
			res := cmd.Main(tc.dir, tc.startHeight, tc.endHeight)
			fmt.Println(res)
		})
	}
}
