package hash

import (
	"hash/fnv"
)

// will compute 2 hashes -> h1 and h2,
func computeHash(data []byte) (uint64, uint64) {
	h := fnv.New64a()

	h.Write(data)
	//make 2 hashes
	h1 := h.Sum64()

	//salting
	h.Write([]byte{1})
	h2 := h.Sum64()

	return h1, h2
}
