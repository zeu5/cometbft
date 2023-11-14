package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	cmtjson "github.com/zeu5/cometbft/libs/json"
	"github.com/zeu5/cometbft/privval"
)

// GenValidatorCmd allows the generation of a keypair for a
// validator.
var GenValidatorCmd = &cobra.Command{
	Use:     "gen-validator",
	Aliases: []string{"gen_validator"},
	Short:   "Generate new validator keypair",
	Run:     genValidator,
}

func genValidator(*cobra.Command, []string) {
	pv := privval.GenFilePV("", "")
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		panic(err)
	}
	fmt.Printf(`%v
`, string(jsbz))
}
