package utils_test

import (
	"encoding/binary"
	"hash/fnv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Hash", func() {
	It("should hash like hash/fnv", func() {
		for _, data := range []string{"", "a", "foobar", "the quick brown fox"} {
			expected := fnv.New64a()
			_, _ = expected.Write([]byte(data))

			hash := utils.NewHash()
			n, err := hash.Write([]byte(data))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(hash.Sum64()).To(Equal(expected.Sum64()))
			Expect(hash.Sum([]byte{1})).To(Equal(expected.Sum([]byte{1})))
		}
	})

	It("should report the size and block size of hash/fnv", func() {
		expected := fnv.New64a()
		hash := utils.NewHash()
		Expect(hash.Size()).To(Equal(expected.Size()))
		Expect(hash.BlockSize()).To(Equal(expected.BlockSize()))
	})

	It("should return to the hash of no data on reset", func() {
		hash := utils.NewHash()
		_, _ = hash.Write([]byte("foobar"))
		hash.Reset()
		Expect(hash).To(Equal(utils.NewHash()))
	})

	It("should continue a copy on its own", func() {
		prefix := utils.NewHash()
		_, _ = prefix.Write([]byte("foo"))
		saved := prefix

		continued := prefix
		_, _ = continued.Write([]byte("bar"))
		Expect(prefix).To(Equal(saved))

		whole := utils.NewHash()
		_, _ = whole.Write([]byte("foobar"))
		Expect(continued).To(Equal(whole))
	})

	It("should write integers in the byte order of the machine", func() {
		expected := utils.NewHash()
		_, _ = expected.Write(binary.NativeEndian.AppendUint64(nil, 0x0102030405060708))
		_, _ = expected.Write(binary.NativeEndian.AppendUint16(nil, 0xfffe))

		hash := utils.NewHash()
		utils.WriteHash(&hash, uint64(0x0102030405060708))
		utils.WriteHash(&hash, int16(-2))
		Expect(hash).To(Equal(expected))
	})

	It("should write a slice like its values one by one", func() {
		values := []int{1, -2, 0x0102030405060708}
		expected := utils.NewHash()
		for _, value := range values {
			utils.WriteHash(&expected, value)
		}

		hash := utils.NewHash()
		utils.WriteHashSlice(&hash, values)
		Expect(hash).To(Equal(expected))
	})

	It("should not change the hash for an empty slice", func() {
		hash := utils.NewHash()
		utils.WriteHashSlice(&hash, []int(nil))
		Expect(hash).To(Equal(utils.NewHash()))
	})
})
