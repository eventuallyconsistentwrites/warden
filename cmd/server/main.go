package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eventuallyconsistentwrites/warden/internal/bloom"
	"github.com/eventuallyconsistentwrites/warden/internal/store"
	"github.com/eventuallyconsistentwrites/warden/shield"
)

// server to handle requests, will init the bloom filter and load items into db and then from db into bloom filter
func main() {
	const (
		dbPath            = "./warden.db"
		tableName         = "users"
		totalUsers        = 1_000_000 //1 million users
		falsePositiveRate = 0.01
		serverPort        = ":8080"
	)

	//init db
	log.Println("Connecting to db...")
	db, err := store.New(dbPath)
	if err != nil {
		fmt.Printf("Error while connecting to db: %v", err)
	}

	log.Println("Connection established!")

	//seed db with 1mil entries
	//create table and fill with 1mil rows if empty
	log.Println("Checking data seed...")
	if err = db.CreateTable(tableName); err != nil {
		fmt.Printf("Error while creating table: %v", err)
	}

	startSeed := time.Now()
	if err = db.Seed(totalUsers, tableName); err != nil {
		log.Fatalf("Failed to seed DB: %v", err)
	}
	fmt.Printf("DB ready! seeding took %v\n", time.Since(startSeed))

	log.Println("Init Bloom filter...")
	bf := bloom.New(uint64(totalUsers), falsePositiveRate)
	log.Println("Bloom filter initialized!")
	log.Println("Loading items from db into bloom filter...")
	startLoad := time.Now()
	count := 0
	err = db.IterateAll(tableName, func(id string) {
		count++
		bf.Add([]byte(id))
		if count%100_000 == 0 {
			fmt.Printf("\r -> Loaded %d items...", count)
		}
	})
	if err != nil {
		log.Fatalf("\nFailed to iterate DB: %v", err)
	}
	fmt.Printf("\n Bloom filter loaded! (%d items in %v)\n", count, time.Since(startLoad))

	warden := shield.New(bf, db, tableName)
	http.HandleFunc("/check", warden.Handler)

	log.Printf("Warden is active on %s/check?id=...", serverPort)
	if err = http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatal(err)
	}

}
