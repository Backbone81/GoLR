package runtime_test

import (
	"io"
	"math/rand"
	"testing"
	"unicode/utf8"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golr/pkg/runtime"
)

var _ = Describe("UTF8RuneReader", func() {
	It("should return the correct sequence of ASCII runes", func() {
		reader := runtime.NewUTF8RuneReader([]byte("abcd"))
		tests := []struct {
			wantRune       rune
			wantRuneSize   int
			wantByteOffset int
			wantLine       int
			wantColumn     int
		}{
			{'a', 1, 0, 1, 1},
			{'b', 1, 1, 1, 2},
			{'c', 1, 2, 1, 3},
			{'d', 1, 3, 1, 4},
		}
		for _, test := range tests {
			Expect(reader.Next()).To(BeTrue())
			Expect(reader.Err()).To(Succeed())

			Expect(reader.Rune()).To(Equal(test.wantRune))
			Expect(reader.RuneSize()).To(Equal(test.wantRuneSize))
			Expect(reader.ByteOffset()).To(Equal(test.wantByteOffset))
			Expect(reader.Line()).To(Equal(test.wantLine))
			Expect(reader.Column()).To(Equal(test.wantColumn))
		}

		// after all runes are read, we need to get the correct error
		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
		Expect(reader.Rune()).To(Equal(utf8.RuneError))
	})

	It("should return the correct sequence of ASCII runes with linebreak", func() {
		reader := runtime.NewUTF8RuneReader([]byte("abc\nxyz"))
		tests := []struct {
			wantRune       rune
			wantRuneSize   int
			wantByteOffset int
			wantLine       int
			wantColumn     int
		}{
			{'a', 1, 0, 1, 1},
			{'b', 1, 1, 1, 2},
			{'c', 1, 2, 1, 3},
			{'\n', 1, 3, 1, 4},
			{'x', 1, 4, 2, 1},
			{'y', 1, 5, 2, 2},
			{'z', 1, 6, 2, 3},
		}
		for _, test := range tests {
			Expect(reader.Next()).To(BeTrue())
			Expect(reader.Err()).To(Succeed())

			Expect(reader.Rune()).To(Equal(test.wantRune))
			Expect(reader.RuneSize()).To(Equal(test.wantRuneSize))
			Expect(reader.ByteOffset()).To(Equal(test.wantByteOffset))
			Expect(reader.Line()).To(Equal(test.wantLine))
			Expect(reader.Column()).To(Equal(test.wantColumn))
		}

		// after all runes are read, we need to get the correct error
		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
		Expect(reader.Rune()).To(Equal(utf8.RuneError))
		Expect(reader.ByteOffset()).To(Equal(tests[len(tests)-1].wantByteOffset + tests[len(tests)-1].wantRuneSize))
	})

	It("should return the correct sequence of ASCII runes with multiple linebreaks", func() {
		reader := runtime.NewUTF8RuneReader([]byte("abc\n\nxyz"))
		tests := []struct {
			wantRune       rune
			wantRuneSize   int
			wantByteOffset int
			wantLine       int
			wantColumn     int
		}{
			{'a', 1, 0, 1, 1},
			{'b', 1, 1, 1, 2},
			{'c', 1, 2, 1, 3},
			{'\n', 1, 3, 1, 4},
			{'\n', 1, 4, 2, 1},
			{'x', 1, 5, 3, 1},
			{'y', 1, 6, 3, 2},
			{'z', 1, 7, 3, 3},
		}
		for _, test := range tests {
			Expect(reader.Next()).To(BeTrue())
			Expect(reader.Err()).To(Succeed())

			Expect(reader.Rune()).To(Equal(test.wantRune))
			Expect(reader.RuneSize()).To(Equal(test.wantRuneSize))
			Expect(reader.ByteOffset()).To(Equal(test.wantByteOffset))
			Expect(reader.Line()).To(Equal(test.wantLine))
			Expect(reader.Column()).To(Equal(test.wantColumn))
		}

		// after all runes are read, we need to get the correct error
		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
		Expect(reader.Rune()).To(Equal(utf8.RuneError))
		Expect(reader.ByteOffset()).To(Equal(tests[len(tests)-1].wantByteOffset + tests[len(tests)-1].wantRuneSize))
	})

	It("should return the correct sequence of ASCII runes with linebreak at the start", func() {
		reader := runtime.NewUTF8RuneReader([]byte("\nabc"))
		tests := []struct {
			wantRune       rune
			wantRuneSize   int
			wantByteOffset int
			wantLine       int
			wantColumn     int
		}{
			{'\n', 1, 0, 1, 1},
			{'a', 1, 1, 2, 1},
			{'b', 1, 2, 2, 2},
			{'c', 1, 3, 2, 3},
		}
		for _, test := range tests {
			Expect(reader.Next()).To(BeTrue())
			Expect(reader.Err()).To(Succeed())

			Expect(reader.Rune()).To(Equal(test.wantRune))
			Expect(reader.RuneSize()).To(Equal(test.wantRuneSize))
			Expect(reader.ByteOffset()).To(Equal(test.wantByteOffset))
			Expect(reader.Line()).To(Equal(test.wantLine))
			Expect(reader.Column()).To(Equal(test.wantColumn))
		}

		// after all runes are read, we need to get the correct error
		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
		Expect(reader.Rune()).To(Equal(utf8.RuneError))
		Expect(reader.ByteOffset()).To(Equal(tests[len(tests)-1].wantByteOffset + tests[len(tests)-1].wantRuneSize))
	})

	It("should return the correct sequence of multibyte UTF-8 runes", func() {
		reader := runtime.NewUTF8RuneReader([]byte("abc£xyz"))
		tests := []struct {
			wantRune       rune
			wantRuneSize   int
			wantByteOffset int
			wantLine       int
			wantColumn     int
		}{
			{'a', 1, 0, 1, 1},
			{'b', 1, 1, 1, 2},
			{'c', 1, 2, 1, 3},
			{'£', 2, 3, 1, 4},
			{'x', 1, 5, 1, 5},
			{'y', 1, 6, 1, 6},
			{'z', 1, 7, 1, 7},
		}
		for _, test := range tests {
			Expect(reader.Next()).To(BeTrue())
			Expect(reader.Err()).To(Succeed())

			Expect(reader.Rune()).To(Equal(test.wantRune))
			Expect(reader.RuneSize()).To(Equal(test.wantRuneSize))
			Expect(reader.ByteOffset()).To(Equal(test.wantByteOffset))
			Expect(reader.Line()).To(Equal(test.wantLine))
			Expect(reader.Column()).To(Equal(test.wantColumn))
		}

		// after all runes are read, we need to get the correct error
		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
		Expect(reader.Rune()).To(Equal(utf8.RuneError))
		Expect(reader.ByteOffset()).To(Equal(tests[len(tests)-1].wantByteOffset + tests[len(tests)-1].wantRuneSize))
	})

	It("should return the correct lexeme", func() {
		reader := runtime.NewUTF8RuneReader([]byte("abcd"))
		Expect(reader.Lexeme(1, 3)).To(Equal([]byte("bc")))
	})

	It("should correctly handle invalid UTF-8 encodings", func() {
		source := []byte{'a', 'b', utf8.RuneSelf + 1, 'c', 'd'}
		reader := runtime.NewUTF8RuneReader(source)
		tests := []struct {
			wantRune       rune
			wantRuneSize   int
			wantByteOffset int
			wantLine       int
			wantColumn     int
		}{
			{'a', 1, 0, 1, 1},
			{'b', 1, 1, 1, 2},
			{utf8.RuneError, 1, 2, 1, 3},
			{'c', 1, 3, 1, 4},
			{'d', 1, 4, 1, 5},
		}
		for i, test := range tests {
			Expect(reader.Next()).To(BeTrue())
			if i == 2 {
				Expect(reader.Err()).To(MatchError(runtime.ErrInvalidUTF8Encoding))
			} else {
				Expect(reader.Err()).To(Succeed())
			}

			Expect(reader.Rune()).To(Equal(test.wantRune))
			Expect(reader.RuneSize()).To(Equal(test.wantRuneSize))
			Expect(reader.ByteOffset()).To(Equal(test.wantByteOffset))
			Expect(reader.Line()).To(Equal(test.wantLine))
			Expect(reader.Column()).To(Equal(test.wantColumn))
		}

		// after all runes are read, we need to get the correct error
		Expect(reader.Next()).To(BeFalse())
		Expect(reader.Err()).To(MatchError(io.EOF))
		Expect(reader.Rune()).To(Equal(utf8.RuneError))
	})
})

func createSingleByteRunes(count int, seed int64) []byte {
	runeSize := 1
	runes := []byte("abcdefghijklmnopqrstuvwxyz")
	if len(runes) != 26*runeSize {
		panic("buffer size mismatch")
	}

	random := rand.New(rand.NewSource(seed))
	result := make([]byte, count*runeSize)
	for i := 0; i < count*runeSize; i += runeSize {
		result[i] = runes[random.Intn(len(runes))]
	}
	return result
}

func createDoubleByteRunes(count int, seed int64) []byte {
	runeSize := 2
	runes := []byte("ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙ")
	if len(runes) != 26*runeSize {
		panic("buffer size mismatch")
	}

	random := rand.New(rand.NewSource(seed))
	result := make([]byte, count*runeSize)
	for i := 0; i < count*runeSize; i += runeSize {
		offset := random.Intn(len(runes)/runeSize) * runeSize
		result[i] = runes[offset]
		result[i+1] = runes[offset+1]
	}
	return result
}

func createTripleByteRunes(count int, seed int64) []byte {
	runeSize := 3
	runes := []byte("ﭐﭑﭒﭓﭔﭕﭖﭗﭘﭙﭚﭛﭜﭝﭞﭟﭠﭡﭢﭣﭤﭥﭦﭧﭨﭩ")
	if len(runes) != 26*runeSize {
		panic("buffer size mismatch")
	}

	random := rand.New(rand.NewSource(seed))
	result := make([]byte, count*runeSize)
	for i := 0; i < count*runeSize; i += runeSize {
		offset := random.Intn(len(runes)/runeSize) * runeSize
		result[i] = runes[offset]
		result[i+1] = runes[offset+1]
		result[i+2] = runes[offset+2]
	}
	return result
}

func createQuadByteRunes(count int, seed int64) []byte {
	runeSize := 4
	runes := []byte("\U00010300\U00010301\U00010302\U00010303\U00010304\U00010305\U00010306\U00010307\U00010308\U00010309\U0001030a\U0001030b\U0001030c\U0001030d\U0001030e\U0001030f\U00010310\U00010311\U00010312\U00010313\U00010314\U00010315\U00010316\U00010317\U00010316\U00010319")
	if len(runes) != 26*runeSize {
		panic("buffer size mismatch")
	}

	random := rand.New(rand.NewSource(seed))
	result := make([]byte, count*runeSize)
	for i := 0; i < count*runeSize; i += runeSize {
		offset := random.Intn(len(runes)/runeSize) * runeSize
		result[i] = runes[offset]
		result[i+1] = runes[offset+1]
		result[i+2] = runes[offset+2]
		result[i+3] = runes[offset+3]
	}
	return result
}

func BenchmarkUTF8RuneReader(b *testing.B) {
	b.Run("1K single byte runes", func(b *testing.B) {
		count := 1024
		runes := createSingleByteRunes(count, 837475)
		b.ResetTimer()
		for range b.N {
			reader := runtime.NewUTF8RuneReader(runes)
			for range count {
				reader.Next()
			}
		}
	})

	b.Run("1K double byte runes", func(b *testing.B) {
		count := 1024
		runes := createDoubleByteRunes(count, 2134652)
		b.ResetTimer()
		for range b.N {
			reader := runtime.NewUTF8RuneReader(runes)
			for range count {
				reader.Next()
			}
		}
	})

	b.Run("1K triple byte runes", func(b *testing.B) {
		count := 1024
		runes := createTripleByteRunes(count, 986798)
		b.ResetTimer()
		for range b.N {
			reader := runtime.NewUTF8RuneReader(runes)
			for range count {
				reader.Next()
			}
		}
	})

	b.Run("1K quad byte runes", func(b *testing.B) {
		count := 1024
		runes := createQuadByteRunes(count, 195735)
		b.ResetTimer()
		for range b.N {
			reader := runtime.NewUTF8RuneReader(runes)
			for range count {
				reader.Next()
			}
		}
	})
}
