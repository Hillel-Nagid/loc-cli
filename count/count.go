package count

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"unicode/utf8"
)

type Result struct {
	Count int
	Files []string
	Dirs  []string
}

func CountLines(target *Result, args ...any) error {
	done := make(chan struct{})
	ctx, cancel := context.WithCancelCause(context.Background())
	var result Result
	if len(args) != 3 {
		return fmt.Errorf("not enough arguments")
	}
	dir, ok := args[0].(*string)
	if !ok {
		return fmt.Errorf("first argument must be a string")
	}
	ignore, ok := args[1].(*string)
	if !ok {
		return fmt.Errorf("second argument must be a string")
	}
	recursive, ok := args[2].(*bool)
	if !ok {
		return fmt.Errorf("third argument must be a boolean")
	}
	ignoreList := strings.Split(strings.ReplaceAll(*ignore, " ", ""), ",")
	go dirLineCounter(ctx, cancel, true, *dir, ignoreList, *recursive, done, &result)
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.Canceled:
		default:
			log.Fatalln(context.Cause(ctx))
		}
	}
	*target = result
	return nil
}

func dirLineCounter(ctx context.Context, cancel context.CancelCauseFunc, root bool, dir string, ignoreList []string, recursive bool, done chan struct{}, result *Result) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		cancel(err)
	}
	if len(entries) > 1 {
		if len(entries) <= 4 {
			// fmt.Println("started processing few")
			entryDoneCh := make(chan struct{}, 1)
			for _, entry := range entries {
				go proceessEntry(ctx, cancel, dir, ignoreList, recursive, entryDoneCh, result, entry)
			}
			for i := 0; i < len(entries); i++ {
				<-entryDoneCh
			}
			// fmt.Println("finished processing few")
		} else {
			// fmt.Println("started processing splits")
			mod := len(entries) % 4
			div := len(entries) / 4
			var splits [][]fs.DirEntry
			for i := 0; i < 4; i++ {
				var addition int
				if i < mod {
					addition = 1
				}
				splits = append(splits, entries[i*div:(i+1)*div+addition])
			}
			splitsDoneCh := make(chan struct{}, 1)
			for _, split := range splits {
				go processEntries(ctx, cancel, dir, ignoreList, recursive, splitsDoneCh, result, split)
			}
			// fmt.Printf("each split is %v entries:\n1. %v\n2. %v\n3. %v\n4. %v\n", div, splits[0], splits[1], splits[2], splits[3])
			for i := 0; i < 4; i++ {
				<-splitsDoneCh
			}
		}
		// fmt.Println("finished processing splits")
	} else if len(entries) == 1 {
		// fmt.Println("started processing single")
		processEntries(ctx, cancel, dir, ignoreList, recursive, done, result, entries)
		<-done
		// fmt.Println("finished processing single")
	}
	if root {
		cancel(nil)
	}
}
func proceessEntry(ctx context.Context, cancel context.CancelCauseFunc, dir string, ignoreList []string, recursive bool, done chan struct{}, result *Result, entry fs.DirEntry) {
	defer func() { done <- struct{}{} }()
	if entry.IsDir() {
		if strings.HasPrefix(entry.Name(), ".") || slices.Contains(ignoreList, entry.Name()) {
			return
		}
		fmt.Printf("Reading directory %s\n", entry.Name())
		dirLineCounter(ctx, cancel, false, path.Join(dir, entry.Name()), ignoreList, recursive, done, result)
		result.Dirs = append(result.Dirs, entry.Name())
	} else {
		fmt.Printf("Reading file %s...\n", path.Join(dir, entry.Name()))
		if fileContent, err := os.ReadFile(path.Join(dir, entry.Name())); err != nil {
			cancel(err)
			return
		} else {
			if utf8.Valid(fileContent) {
				fileString := string(fileContent)
				contentLines := strings.Split(fileString, "\n")
				result.Files = append(result.Files, entry.Name())
				result.Count += len(contentLines)
			}
		}
	}
	return
}
func processEntries(ctx context.Context, cancel context.CancelCauseFunc, dir string, ignoreList []string, recursive bool, done chan struct{}, result *Result, entries []fs.DirEntry) {
	entryDoneCh := make(chan struct{}, 1)
	for _, entry := range entries {
		proceessEntry(ctx, cancel, dir, ignoreList, recursive, entryDoneCh, result, entry)
		<-entryDoneCh
	}
	done <- struct{}{}
}
