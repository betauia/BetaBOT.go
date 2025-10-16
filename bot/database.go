package bot

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDatabase() {
	var err error
	DB, err = sql.Open("sqlite3", "./db/beta-scheduled_messages.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create the scheduled_messages table if it doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS scheduled_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		guild_id TEXT NOT NULL,
		role_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		message TEXT NOT NULL,
		scheduled_time DATETIME NOT NULL,
		channel_id TEXT NOT NULL
	);
	`
	_, err = DB.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Database initialized successfully")
}
