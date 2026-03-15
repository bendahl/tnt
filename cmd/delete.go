package cmd

import (
	"flag"
	"fmt"
	"log"

	"github.com/bendahl/tnt/cmd/util"
)

func Delete(args []string) {
	var team string
	dfs := flag.NewFlagSet("delete", flag.ExitOnError)
	dfs.StringVar(&team, "t", "", "name of the team")
	err := dfs.Parse(args)
	if err != nil {
		log.Fatalf("failed to parse arguments: %v", err)
	}
	if team == "" {
		log.Fatalf("no team given")
	}
	_, err = util.Kubectl(fmt.Sprintf("delete ns -l 'team in (%s)'", team))
	if err != nil {
		log.Fatalf("failed to delete namespace: %v", err)
	}
	_, err = util.Kubectl(fmt.Sprintf("delete clusterrolebinding -l 'team in (%s)'", team))
	if err != nil {
		log.Fatalf("failed to delete clusterrolebinding: %v", err)
	}
	_, err = util.Kubectl(fmt.Sprintf("delete serviceaccount -n kube-system -l 'team in (%s)'", team))
	if err != nil {
		log.Fatalf("failed to delete serviceaccount: %v", err)
	}
}
