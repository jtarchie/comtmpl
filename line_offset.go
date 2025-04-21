package main

import (
	"bufio"
	"os"
	"sort"
)

type LineIndex struct {
	NewlineOffsets []int64
}

// Build a line offset index
func NewLineIndex(filePath string) (*LineIndex, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var newlineOffsets []int64
	var offset int64 = 0

	reader := bufio.NewReader(file)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break // EOF is expected
		}

		offset++
		if b == '\n' {
			newlineOffsets = append(newlineOffsets, offset)
		}
	}

	return &LineIndex{NewlineOffsets: newlineOffsets}, nil
}

// Get line number at given byte offset
func (li *LineIndex) LineNumberAt(offset int64) int {
	// Binary search: count how many newline offsets are before the given offset
	idx := sort.Search(len(li.NewlineOffsets), func(i int) bool {
		return li.NewlineOffsets[i] > offset
	})

	return idx + 1 // lines are 1-based
}
