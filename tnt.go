package main

import (
	"fmt"
	"os"

	"github.com/bendahl/tnt/cmd"
)

var commands = map[string]string{
	"create": "Create a new team namespace",
	"delete": "Delete an existing team namespace",
	"list":   "List existing team namespaces",
}

func main() {
	args := os.Args[1:]
	numArgs := len(args)

	if numArgs < 1 {
		fmt.Printf("Expected at least two arguments, got %v\n", numArgs)
		fmt.Println()
		usage()
		os.Exit(1)
	}

	command := args[0]
	if !isValidCommand(command) {
		fmt.Printf("Unknown command \"%s\"\n", command)
		fmt.Println()
		usage()
		os.Exit(1)
	}

	switch command {
	case "create":
		cmd.Create(args[1:])
	case "delete":
		cmd.Delete(args[1:])
	case "list":
		cmd.List(args[1:])
	default:
		panic("ERROR: Not implemented - This should be unreachable")
	}
}

func isValidCommand(cmd string) bool {
	_, ok := commands[cmd]
	return ok
}

func usage() {
	fmt.Println("Usage: tnt <command> <args>")
	fmt.Println()
	fmt.Println("Available commands:")
	for c, description := range commands {
		fmt.Printf("	%s - %s\n", c, description)
	}
}
