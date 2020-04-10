/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package checkfile

// TermsBuffer holds a buffer for the file and the search terms.
type TermsBuffer struct {
	Buffer    []byte
	Terms     [][]byte
	MinLength int
	MaxLength int
	Unmatched []int
}

// NewTermsBuffer creates a new instance of TermsBuffer.
func NewTermsBuffer(bufferSize int, searchTerms []string) *TermsBuffer {
	termsBuffer := new(TermsBuffer)
	termsBuffer.SetTerms(searchTerms)
	termsBuffer.initBuffer(bufferSize)
	return termsBuffer
}

// NewTermsBufferFromBytes creates a new instance of TermsBuffer.
func NewTermsBufferFromBytes(bufferSize int, searchTerms [][]byte) *TermsBuffer {
	termsBuffer := new(TermsBuffer)
	termsBuffer.Terms = searchTerms
	termsBuffer.initMinMax()
	termsBuffer.initUnmatched()
	termsBuffer.initBuffer(bufferSize)
	return termsBuffer
}

// SetTerms sets the search terms and the minimum and maximum term length.
// Empty search terms are removed.
func (termsBuffer *TermsBuffer) SetTerms(searchTerms []string) {
	termsBuffer.initTerms(searchTerms)
	termsBuffer.initMinMax()
	termsBuffer.initUnmatched()
}

// SetTermsFromBytes sets the search terms and the minimum and maximum term length.
// Empty search terms are not removed.
func (termsBuffer *TermsBuffer) SetTermsFromBytes(searchTerms [][]byte) {
	termsBuffer.Terms = searchTerms
	termsBuffer.initMinMax()
	termsBuffer.initUnmatched()
}

func (termsBuffer *TermsBuffer) initTerms(searchTerms []string) {
	bytes := make([][]byte, 0, len(searchTerms))
	for _, term := range searchTerms {
		if len(term) > 0 {
			bytes = append(bytes, []byte(term))
		}
	}
	termsBuffer.Terms = bytes
}

func (termsBuffer *TermsBuffer) initMinMax() {
	termsBuffer.MinLength = int(^uint(0) >> 1)
	termsBuffer.MaxLength = 0

	for _, term := range termsBuffer.Terms {
		length := len(term)

		if length < termsBuffer.MinLength {
			termsBuffer.MinLength = length
		}
		if length > termsBuffer.MaxLength {
			termsBuffer.MaxLength = length
		}
	}
	if termsBuffer.MaxLength == 0 {
		termsBuffer.MinLength = 0
	}
}

func (termsBuffer *TermsBuffer) initUnmatched() {
	minLength := len(termsBuffer.Terms)

	if cap(termsBuffer.Unmatched) < minLength {
		termsBuffer.Unmatched = make([]int, minLength)

	} else {
		termsBuffer.Unmatched = termsBuffer.Unmatched[:minLength]
	}
	for i := range termsBuffer.Unmatched {
		termsBuffer.Unmatched[i] = i
	}
}

func (termsBuffer *TermsBuffer) initBuffer(bufferSize int) {
	size := max(bufferSize, termsBuffer.minBufferSize())
	termsBuffer.Buffer = make([]byte, size)
}

func (termsBuffer *TermsBuffer) minBufferSize() int {
	return int(float64(termsBuffer.MaxLength) * 1.2)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
