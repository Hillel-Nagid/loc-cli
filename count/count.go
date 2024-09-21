package count

import (
	"context"
	"fmt"
	"io/fs"
	"loc-cli/filetree"
	"loc-cli/utils"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"time"
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
	dir, recursive, includeBlanks, includeComments, ignoreList, err := parseArgs(args)
	if err != nil {
		return err
	}
	tree, err := filetree.NewFileTree(*dir, ignoreList)
	if err != nil {
		return err
	}
	go dirLineCounter(ctx, cancel, true, *dir, ignoreList, *recursive, *includeBlanks, *includeComments, done, &result, tree)
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

func dirLineCounter(ctx context.Context, cancel context.CancelCauseFunc, root bool, dir string, ignoreList []string, recursive bool, includeBlanks bool, includeComments bool, done chan struct{}, result *Result, tree *filetree.FileTree) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		cancel(err)
	}
	if len(entries) > 1 {
		if len(entries) <= 4 {
			entryDoneCh := make(chan struct{}, 1)
			for _, entry := range entries {
				go processEntry(ctx, cancel, dir, ignoreList, recursive, includeBlanks, includeComments, entryDoneCh, result, entry, tree)
			}
			for i := 0; i < len(entries); i++ {
				<-entryDoneCh
			}
		} else {
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
				go processEntries(ctx, cancel, dir, ignoreList, recursive, includeBlanks, includeComments, splitsDoneCh, result, split, tree)
			}
			for i := 0; i < 4; i++ {
				<-splitsDoneCh
			}
		}
	} else if len(entries) == 1 {
		processEntry(ctx, cancel, dir, ignoreList, recursive, includeBlanks, includeComments, done, result, entries[0], tree)
		<-done
	}
	if root {
		cancel(nil)
	}
}
func processEntry(ctx context.Context, cancel context.CancelCauseFunc, dir string, ignoreList []string, recursive bool, includeBlanks bool, includeComments bool, done chan struct{}, result *Result, entry fs.DirEntry, tree *filetree.FileTree) {
	defer func() { done <- struct{}{} }()
	entryPath := path.Join(dir, entry.Name())
	if entry.IsDir() {
		if strings.HasPrefix(entry.Name(), ".") || slices.Contains(ignoreList, entry.Name()) {
			return
		}
		dirLineCounter(ctx, cancel, false, entryPath, ignoreList, recursive, includeBlanks, includeComments, done, result, tree)
		result.Dirs = append(result.Dirs, entry.Name())
	} else {
		if strings.HasSuffix(entry.Name(), ".sum") || slices.Contains(ignoreList, entry.Name()) {
			return
		}
		tree.Loading(len(result.Files))
		time.Sleep(time.Millisecond * 200)
		if fileContent, err := os.ReadFile(entryPath); err != nil {
			cancel(err)
			return
		} else {
			if utf8.Valid(fileContent) {
				fileString := string(fileContent)
				contentLines := utils.Filter(strings.Split(fileString, "\n"), func(s string) bool {
					if !includeBlanks && s == "" {
						return false
					}
					if !includeComments && (strings.HasPrefix(s, "//") || strings.HasPrefix(s, "/*") || strings.HasSuffix(s, "*/")) {
						return false
					}
					return true
				})
				tree.ChangeFileStatus(entryPath, filetree.DoneEntryStatus)
				result.Files = append(result.Files, entry.Name())
				result.Count += len(contentLines)
			}
		}
	}
}
func processEntries(ctx context.Context, cancel context.CancelCauseFunc, dir string, ignoreList []string, recursive bool, includeBlanks bool, includeComments bool, done chan struct{}, result *Result, entries []fs.DirEntry, tree *filetree.FileTree) {
	entryDoneCh := make(chan struct{}, 1)
	for _, entry := range entries {
		processEntry(ctx, cancel, dir, ignoreList, recursive, includeBlanks, includeComments, entryDoneCh, result, entry, tree)
		<-entryDoneCh
	}
	done <- struct{}{}
}

func parseArgs(args []any) (dir *string, recursive *bool, includeBlanks *bool, includeComments *bool, ignoreList []string, err error) {
	if len(args) != 5 {
		err = fmt.Errorf("not enough arguments")
	}
	dir, ok := args[0].(*string)
	if !ok {
		err = fmt.Errorf("first argument must be a string")
	}
	ignore, ok := args[1].(*string)
	if !ok {
		err = fmt.Errorf("second argument must be a string")
	}
	recursive, ok = args[2].(*bool)
	if !ok {
		err = fmt.Errorf("third argument must be a boolean")
	}
	includeBlanks, ok = args[3].(*bool)
	if !ok {
		err = fmt.Errorf("fourth argument must be a boolean")
	}
	includeComments, ok = args[4].(*bool)
	if !ok {
		err = fmt.Errorf("fifth argument must be a boolean")
	}
	ignoreList = strings.Split(strings.ReplaceAll(*ignore, " ", ""), ",")
	return
}
