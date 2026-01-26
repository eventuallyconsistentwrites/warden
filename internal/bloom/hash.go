package bloom

import (
	"hash/fnv"
)

// will compute 2 hashes -> h1 and h2,
func computeHash(data []byte) (uint64, uint64) {
	h := fnv.New64a()

	h.Write(data)
	//make 2 hashes
	h1 := h.Sum64()

	h.Reset()

	h.Write(data)
	//salting
	//this is basically adding something to the same hash to get some completely diff value
	h.Write([]byte{1})
	h2 := h.Sum64()

	return h1, h2
}
