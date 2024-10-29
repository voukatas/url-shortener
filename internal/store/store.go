package store

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	Db *sql.DB
}

type Store interface {
	Shorten(string) (int64, error)
	Lookup(int64) (string, error)
	Close()
	SetStoreOptions()
}

func (d *DB) Close() {
	d.Db.Close()
}

func (d *DB) SetStoreOptions() {
	d.Db.SetMaxOpenConns(10)
	d.Db.SetMaxIdleConns(5)

}

func NewStore(dbPath string) (Store, error) {
	fmt.Println("Attempting to open database...")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	fmt.Println("Database opened!")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Short_Url_Service (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		Long_url TEXT NOT NULL
	)`)
	if err != nil {
		return nil, err
	}
	// Enable WAL mode to allow for concurrent reads and a single write
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, err
	}

	fmt.Println("Table created or exists!")

	return &DB{Db: db}, nil
}

func (d *DB) Shorten(longUrl string) (int64, error) {
	result, err := d.Db.Exec(`INSERT INTO Short_Url_Service (Long_url) VALUES (?)`, longUrl)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DB) Lookup(shortCode int64) (string, error) {
	var originalURL string
	err := d.Db.QueryRow(`SELECT Long_url FROM Short_Url_Service WHERE  ID = ?`, shortCode).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("short URL not found")
		}
		return "", err
	}
	return originalURL, nil
}
