package cmd

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/bendahl/tnt/cmd/util"
)

func List(args []string) {
	var team string
	var outFormat string
	lfs := flag.NewFlagSet("list", flag.ExitOnError)
	lfs.StringVar(&team, "t", "", "name of the team")
	lfs.StringVar(&outFormat, "o", "", "use a different output format (e.g. -o=yaml)")
	err := lfs.Parse(args)
	if err != nil {
		log.Fatalf("failed to parse arguments: %v", err)
	}
	cmdBuilder := new(strings.Builder)
	cmdBuilder.WriteString("get ns -l team")
	if len(team) > 0 {
		cmdBuilder.WriteString(fmt.Sprintf(" -l 'team in (%s)'", team))
	}

	if len(outFormat) > 0 {
		cmdBuilder.WriteString(fmt.Sprintf(" -o=%s", outFormat))
	}
	msg, err := util.Kubectl(cmdBuilder.String())
	if err != nil {
		log.Fatalf("listing team namespaces failed: %v, msg: %v", err, msg)
	}
	fmt.Println(msg)
}
