package internal

import (
	"time"

	"github.com/boltdb/bolt"
)

func writeToDB() error {
	db, err := bolt.Open("./db/my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	defer db.Close()

	return nil
}
