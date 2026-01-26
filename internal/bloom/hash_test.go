package bloom

import (
	"testing"
)

func TestComputeHash_Determinisitic(t *testing.T) {
	data := []byte("Hello World!")
	h1a, h2a := computeHash(data)
	h1b, h2b := computeHash(data)

	t.Logf("First run h1: %d, h2:%d", h1a, h2a)
	t.Logf("Second run h1: %d, h2:%d", h1b, h2b)

	if h1a != h1b {
		t.Fatalf("h1 not deterministic: %d != %d", h1a, h1b)
	}

	if h2a != h2b {
		t.Fatalf("h2 not determinisitic: %d!=%d", h2a, h2b)
	}
}

func TestComputeHash_DiffInputs(t *testing.T) {
	data1 := []byte("Hello World")
	data2 := []byte("Hello World!")
	h1a, h2a := computeHash(data1)
	h1b, h2b := computeHash(data2)

	t.Logf("First run h1: %d, h2:%d", h1a, h2a)
	t.Logf("Second run h1: %d, h2:%d", h1b, h2b)

	if h1a == h1b {
		t.Fatalf("h1 should be different for different inputs")
	}

	if h2a == h2b {
		t.Fatalf("h2 should be different for different inputs")
	}
}
