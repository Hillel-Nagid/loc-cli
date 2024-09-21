package filetree

import (
	"math"
	"strconv"
)

func createLoader(currFilesCount int, totalFileCount int) string {
	loader := "["
	precentComplete := math.Ceil((float64(currFilesCount)/float64(totalFileCount))*100) / 5
	for i := 0; i < int(precentComplete); i++ {
		loader += "▮"
	}
	for i := 0; i < 20-int(precentComplete); i++ {
		loader += " "
	}
	return loader + "] " + strconv.Itoa(currFilesCount) + "/" + strconv.Itoa(totalFileCount)
}

func padTree(prevDir string) string {
	var pad string
	for i := 0; i < len(prevDir); i++ {
		pad += "_"
	}
	return pad
}
