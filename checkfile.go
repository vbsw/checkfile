/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

// Package checkfile provides functions to check file status and content.
package checkfile

import (
	"github.com/vbsw/slices/remove"
	"io"
	"os"
)

// Exists returns true, if the file exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	exists := err == nil || !os.IsNotExist(err)
	return exists
}

// IsDirectory returns true, if the file exists and is a directory.
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	exists := fileInfo != nil && fileInfo.IsDir() && (err == nil || !os.IsNotExist(err))
	return exists
}

// IsFile returns true, if the file exists and is a file.
func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	exists := fileInfo != nil && !fileInfo.IsDir() && (err == nil || !os.IsNotExist(err))
	return exists
}

// Size returns the size of the file in bytes. If the file does not exist it returns 0.
func Size(path string) int64 {
	var size int64
	fileInfo, err := os.Stat(path)
	exists := fileInfo != nil && !fileInfo.IsDir() && (err == nil || !os.IsNotExist(err))

	if exists {
		size = fileInfo.Size()
	}
	return size
}

// ContainsAll returns true, if the file exists and contains all of the search terms.
func ContainsAll(path string, termsBuffer *SearchTermsBuffer) (bool, error) {
	var containsAll bool
	fileInfo, err := os.Stat(path)

	if fileInfo != nil && (err == nil || !os.IsNotExist(err)) {
		if len(termsBuffer.Terms) > 0 {
			if fileInfo.Size() > 0 {
				var file *os.File
				file, err = os.Open(path)

				if err == nil {
					defer file.Close()
					termsBuffer.initUnmatched()
					containsAll, err = containsAllFromFile(file, termsBuffer)
				}
			} // else containsAll = false
		} else {
			containsAll = true
		}
	}
	return containsAll, err
}

// ContainsAny returns true, if the file exists and contains at least one of the search terms.
func ContainsAny(path string, termsBuffer *SearchTermsBuffer) (bool, error) {
	var containsAny bool
	fileInfo, err := os.Stat(path)

	if fileInfo != nil && (err == nil || !os.IsNotExist(err)) {
		if len(termsBuffer.Terms) > 0 {
			if fileInfo.Size() > 0 {
				var file *os.File
				file, err = os.Open(path)

				if err == nil {
					defer file.Close()
					containsAny, err = containsAnyFromFile(file, termsBuffer)
				}
			} // else containsAny = false
		} else {
			containsAny = true
		}
	}
	return containsAny, err
}

func containsAllFromFile(file *os.File, termsBuffer *SearchTermsBuffer) (bool, error) {
	var err error
	var containsAll bool
	var nRead, nProcessed int

	nRead, err = file.Read(termsBuffer.Buffer)

	for err == nil {
		nProcessed = searchAll(termsBuffer, nRead)
		containsAll = len(termsBuffer.Unmatched) == 0

		if !containsAll && nRead == len(termsBuffer.Buffer) {
			nUnread := len(termsBuffer.Buffer) - nProcessed
			copy(termsBuffer.Buffer, termsBuffer.Buffer[nProcessed:])
			nRead, err = file.Read(termsBuffer.Buffer[nUnread:])
			nRead += nUnread

		} else {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return containsAll, err
}

func searchAll(termsBuffer *SearchTermsBuffer, nRead int) int {
	var nProcessed int
	bufferLimit := nRead - termsBuffer.MinLength + 1

	for nProcessed = 0; nProcessed < bufferLimit && len(termsBuffer.Unmatched) > 0; nProcessed++ {
		for j := 0; j < len(termsBuffer.Unmatched); j++ {
			termIndex := termsBuffer.Unmatched[j]
			term := termsBuffer.Terms[termIndex]

			if len(termsBuffer.Buffer)-nProcessed >= len(term) {
				contains := containsTerm(termsBuffer.Buffer, nProcessed, term)

				if contains {
					termsBuffer.Unmatched = remove.Int(termsBuffer.Unmatched, j)
					j--
				}
			}
		}
	}
	return nProcessed
}

func containsAnyFromFile(file *os.File, termsBuffer *SearchTermsBuffer) (bool, error) {
	var err error
	var containsAny bool
	var nRead, nProcessed int

	nRead, err = file.Read(termsBuffer.Buffer)

	for err == nil {
		nProcessed, containsAny = searchAny(termsBuffer, nRead)

		if !containsAny && nRead == len(termsBuffer.Buffer) {
			nUnread := len(termsBuffer.Buffer) - nProcessed
			copy(termsBuffer.Buffer, termsBuffer.Buffer[nProcessed:])
			nRead, err = file.Read(termsBuffer.Buffer[nUnread:])
			nRead += nUnread

		} else {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return containsAny, err
}

func searchAny(termsBuffer *SearchTermsBuffer, nRead int) (int, bool) {
	var nProcessed int
	var containsAny bool
	bufferLimit := nRead - termsBuffer.MinLength + 1

	for nProcessed = 0; nProcessed < bufferLimit && !containsAny; nProcessed++ {
		for _, term := range termsBuffer.Terms {
			if len(termsBuffer.Buffer)-nProcessed >= len(term) {
				containsAny = containsTerm(termsBuffer.Buffer, nProcessed, term)
				if containsAny {
					break
				}
			}
		}
	}
	return nProcessed, containsAny
}

func containsTerm(buffer []byte, i int, term []byte) bool {
	for k, b := range term {
		if buffer[i+k] != b {
			return false
		}
	}
	return true
}
