package filetree

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type EntryStatus int // Pending = 0 Done = 1 Skipped = 2

const (
	PendingEntryStatus EntryStatus = iota
	DoneEntryStatus
	SkippedEntryStatus
)

type entry struct {
	Name   string
	path   string
	Status EntryStatus
}

type FileTree struct {
	Tree        string
	Directories map[int]*entry
	Files       map[int]*entry
}

func (f *FileTree) Loading(currFilesCount int) {
	fmt.Printf("\r%s", createLoader(currFilesCount, len(f.Files)))
}

func (f *FileTree) GetDirecory(path string) *entry {
	var e *entry
	for id, dir := range f.Directories {
		if dir.path == strings.ReplaceAll(path, "/", "\\") {
			e = f.Directories[id]
		}
	}
	return e
}

func (f *FileTree) GetFile(path string) (*entry, int) {
	var e *entry
	var ID int
	for id, file := range f.Files {
		if file.path == strings.ReplaceAll(path, "/", "\\") {
			e = f.Files[id]
			ID = id
		}
	}
	return e, ID
}

func (f *FileTree) ChangeFileStatus(path string, status EntryStatus) {
	if entry, id := f.GetFile(path); entry != nil {
		var currColor string
		switch entry.Status {
		case DoneEntryStatus:
			currColor = "\033[32m"
		case PendingEntryStatus:
			currColor = "\033[33m"
		case SkippedEntryStatus:
			currColor = "\033[31m"
		}
		entry.Status = status
		var newColor string
		switch status {
		case DoneEntryStatus:
			newColor = "\033[32m"
		case PendingEntryStatus:
			newColor = "\033[33m"
		case SkippedEntryStatus:
			newColor = "\033[31m"
		}
		allLines := strings.Split(f.Tree, "\n")
		tmp := strings.ReplaceAll(allLines[id], currColor, newColor)
		allLines[id] = tmp
		f.Tree = strings.Join(allLines, "\n")
	}
}

func NewFileTree(root string, ignoreList []string) (*FileTree, error) {
	fileTree := &FileTree{
		Directories: map[int]*entry{},
		Files:       map[int]*entry{},
	}
	rootPath := strings.ReplaceAll(root, "/", "\\")
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		path = strings.ReplaceAll(path, rootPath, "")
		ignoreDir, _ := regexp.MatchString(`\.[\w\d\S]*\\.*`, path)
		if !strings.HasSuffix(path, ".sum") && !ignoreDir {
			treeDirs := strings.Split(path, "\\")
			if d.IsDir() {
				if len(treeDirs) == 1 {
					fileTree.Directories[0] = &entry{Name: d.Name(), path: path}
				}
				fileTree.Directories[len(strings.Split(fileTree.Tree, "\n"))] = &entry{Name: d.Name(), path: path}
			} else {
				fileTree.Files[len(strings.Split(fileTree.Tree, "\n"))] = &entry{Name: d.Name(), path: path}
			}
			var treeName string
			if slices.ContainsFunc(ignoreList, func(s string) bool {
				return strings.Contains(path, s)
			}) {
				treeName = " \033[37m" + d.Name() + "\033[0m"
			} else {
				treeName = " \033[31m" + d.Name() + "\033[0m"
			}
			if len(treeDirs) == 2 {
				fileTree.Tree += "\n|" + treeName
			} else if len(treeDirs) == 1 {
				fileTree.Tree += "\n" + treeName
			} else {
				parentDir := strings.Join(treeDirs[:len(treeDirs)-1], "/")
				fileTree.Tree += "\n|" + padTree(parentDir) + treeName
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return fileTree, nil
}
