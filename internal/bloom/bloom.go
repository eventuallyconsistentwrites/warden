package bloom

import (
	"math"
)

type BloomFilter struct {
	bitset []uint64 //represents 1D array for the filter, and each integer can store 64 bits
	k      uint64   //number of hash functions
	m      uint64   //number of bits required, where the bitset size is (m+63)/64
}

// constructor
// n -> number of expected items -> eg: 1 million entries in the db
// p -> probability of error or the false positive rate
func New(n uint64, p float64) *BloomFilter { //size will be number of elements, eg: 1 million entries in the db
	if n <= 0 || p < 0.0 || p > 1.0 {
		panic("Not possible to create a bloom filter with these constraints")
	}
	size := computeBloomFilterSize(n, p) //no of bits total
	k := uint64(float64(size) / float64(n) * math.Log(2))
	return &BloomFilter{
		bitset: make([]uint64, calculateSliceSize(size)),
		k:      k,
		m:      size,
	}
}

func computeBloomFilterSize(n uint64, p float64) uint64 {
	//known formula for calculating size
	//roughly the filter size is ~9.5 times the expected items, if complex we can just substitute that too for 1% error
	return uint64(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2))
}

// TODO: Add item, rn not thread safe as adding while starting the server then not adding
func (b *BloomFilter) Add(item []byte) {
	//will addd using hash functions
	//run it through k hash functions and then add to bitset
	//for each item, hash it k times
	k := b.k
	m := b.m
	bitset := b.bitset
	hash1, hash2 := computeHash(item)
	var i uint64
	for i = 0; i < k; i++ {
		hash := hash1 + (i * hash2)

		//now to add
		//first mod by number of bits
		bitIndex := hash % m //indicates which bit to set

		//in array, each index can store 64 bits
		arrIndex := bitIndex / 64
		offSet := bitIndex % 64

		bitset[arrIndex] = bitset[arrIndex] | (uint64(1) << offSet)

	}
}

// TODO: Check if present
func (b *BloomFilter) Contains(item []byte) bool {
	k := b.k
	m := b.m
	bitset := b.bitset
	hash1, hash2 := computeHash(item)
	var i uint64
	for i = 0; i < k; i++ {
		hash := hash1 + (i * hash2)

		//now to add
		//first mod by number of bits
		bitIndex := hash % m //indicates which bit to set

		//in array, each index can store 64 bits
		arrIndex := bitIndex / 64
		offSet := bitIndex % 64

		num := bitset[arrIndex]
		if num&((uint64(1))<<offSet) == 0 {
			return false //definitely not there
		}
	}
	return true //might be there
}

// Computes the size required for the bitset, each entry can store 64 values
func calculateSliceSize(bitsNeeded uint64) uint64 {
	return (bitsNeeded + 63) / 64
}
