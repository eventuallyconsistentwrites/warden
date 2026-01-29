package bloom

import (
	"fmt"
	"testing"
)

// test cases for bloom ->
// 1. Sanity Check
// 2. Error Rate
func TestBloomFilter_Basic(t *testing.T) {
	bf := New(10, 0.01)

	item1 := []byte("HELLO")
	bf.Add(item1)
	if !bf.Contains(item1) {
		t.Errorf("Critical: Added %v but not found in the filter", string(item1))
	}

	item2 := []byte("WORLD")
	//world should likely not exist. if does,then it is a collision
	if bf.Contains(item2) {
		t.Log("Warning: Collision detected in the basic test")
	}
}

func TestBloomFilter_ErrorRate(t *testing.T) {
	n := uint64(1000)
	p := 0.01
	bf := New(n, p)
	for i := 0; i < 1000; i++ {
		item := []byte(fmt.Sprintf("user_%d", i))
		bf.Add(item)
	}

	//now check error rate against 9k users that do not exist
	//user 1000 - user 9999 dont exist
	trials := 9000
	falsePositives := 0
	for i := 1000; i < trials+1000; i++ {
		fakeItem := []byte(fmt.Sprintf("user_%d", i))

		if bf.Contains(fakeItem) {
			falsePositives++
		}
	}
	rate := float64(falsePositives) / float64(trials)
	if rate > (p * 1.5) {
		t.Errorf("Error rate too high: Expected : ~%f, gott ~%f", p, rate)
	}
	t.Logf("False positive rate : %f (Target: %f)", rate, p)
}
