// will intercept the db requests, verify with the same with bloom filter for existence
package shield

import (
	"fmt"
	"net/http"

	"github.com/eventuallyconsistentwrites/warden/internal/bloom"
	"github.com/eventuallyconsistentwrites/warden/internal/store"
)

// Shield is per-table, in prod systems, it will be enabled for 'hot' resources like specific tables
type Shield struct {
	filter    *bloom.BloomFilter
	store     *store.Store
	tableName string
	isEnabled bool
}

// default isEnabled true since it is a shield
func New(b *bloom.BloomFilter, s *store.Store, name string) *Shield {
	return &Shield{
		filter:    b,
		store:     s,
		tableName: name,
		isEnabled: true,
	}
}

// handler is the http middlware that will intercept requ4sts
func (s *Shield) Handler(w http.ResponseWriter, r *http.Request) {
	//parse id:
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	if s.isEnabled && s.filter != nil {
		//check if present in the bloom filter
		if !s.filter.Contains([]byte(id)) {
			//blocked by shield
			// fmt.Println("Blocked by Bloom Filter:", id)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	}

	//1. Shield enabled and it is present/false positive
	//2. Shield disabled
	exists, err := s.store.Check(id, s.tableName)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "User not found in DB", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("User %s found!", id)))

}

func (s *Shield) Disable() {
	s.isEnabled = false
}

func (s *Shield) Enable() {
	s.isEnabled = true
}
