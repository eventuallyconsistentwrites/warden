//db layer for sqlite db

package store

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("Could not open the db with error: %w", err)
	}

	//ping db to check connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Error while pinging the db: %w", err)
	}
	return &Store{
		db: db,
	}, nil
}

func (s *Store) CreateTable(name string) error {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(id TEXT PRIMARY KEY)", name)
	_, err := s.db.Exec(query)
	return err
}

// seed with n dummy entries
// uses transactions and prepare queries -> without this, doing a lot if inserts would take a very long time
func (s *Store) Seed(n int, name string) error {

	//first we check if already exists
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", name)
	s.db.QueryRow(query).Scan(&count)

	if count >= n {
		fmt.Println("DB already seeded, skipping seeding")
		return nil
	}

	fmt.Printf("Seeding %s with %d entries...\n", name, n)
	//use transactions
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	insertQuery := fmt.Sprintf("INSERT OR IGNORE INTO %s (id) VALUES (?)", name)
	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("user_%d", i)
		if _, err := stmt.Exec(id); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Check(id string, name string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE id = ? LIMIT 1", name)

	var exists int
	err := s.db.QueryRow(query, id).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// loading entire data into memory at once might be too much of an overhead, we save memory using a callback passed
func (s *Store) IterateAll(name string, fn func(id string)) error {
	query := fmt.Sprintf("SELECT id FROM %s", name)
	rows, err := s.db.Query(query) //returns a pointer to the rows
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		fn(id) //callback to pass the id, once passed, this(current) function forgets that row
	}
	return rows.Err()
}
