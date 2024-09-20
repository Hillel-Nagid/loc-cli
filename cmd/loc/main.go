package main

import (
	"flag"
	"fmt"
	"loc-cli/command"
	"loc-cli/count"
	"log"
	"os"
	"strings"
)

func main() {
	countFlag := flag.NewFlagSet("count", flag.ExitOnError)
	repoCountFlag := countFlag.String("repo", ".", "Specify which repository's lines to count. \".\" by default")
	ignoreCountFlag := countFlag.String("ignore", "", "Specify which directories to ignore. empty by default")
	recursiveCountFlag := countFlag.Bool("recursive", true, "Specify wether you wuld like to count lines recursively on the repo, or not. true by default")
	var loc count.Result
	countCommand := command.NewCommand(countFlag, count.CountLines, []any{repoCountFlag, ignoreCountFlag, recursiveCountFlag}, &loc)
	if len(os.Args) < 2 {
		fmt.Println("expected a subcommand")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "count":
		if err := countFlag.Parse(os.Args[2:]); err != nil {
			log.Fatalf("error parsing count command: %v", err)
		}
		fmt.Printf("Counting lines at repo \"%s\"...\n", *repoCountFlag)
		countCommand.Run()
		fmt.Printf("lines count: %v\n", loc.Count)
		fmt.Printf("files: %v\n", strings.Join(loc.Files, ", "))
		fmt.Printf("directories: %v\n", strings.Join(loc.Dirs, ", "))
	}
}
