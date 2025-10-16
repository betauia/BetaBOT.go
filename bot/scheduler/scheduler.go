package scheduler

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/betauia/BetaBot.go/bot/models"
	"github.com/bwmarrin/discordgo"
)

var (
	isRunning bool
	mu        sync.Mutex
)

func Start(session *discordgo.Session, db *sql.DB, checkInterval time.Duration) {
	mu.Lock()
	if isRunning {
		mu.Unlock()
		log.Println("Scheduler is already running")
		return
	}
	isRunning = true
	mu.Unlock()

	go run(session, db, checkInterval)
}

func run(session *discordgo.Session, db *sql.DB, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	log.Printf("✅ Scheduler started - checking every %v", checkInterval)

	for range ticker.C {
		checkAndSend(session, db)
	}
}

// checkAndSend retrieves and sends due messages
func checkAndSend(session *discordgo.Session, db *sql.DB) {
	dueMessages, err := models.GetPendingMessages(db)
	if err != nil {
		log.Printf("❌ Error fetching due messages: %v", err)
		return
	}

	if len(dueMessages) == 0 {
		return
	}

	log.Printf("📬 Found %d message(s) to send", len(dueMessages))

	for _, msg := range dueMessages {
		go sendMessage(session, db, msg)
	}
}

// sendMessage sends a single scheduled message
func sendMessage(session *discordgo.Session, db *sql.DB, msg *models.ScheduledMessage) {
	log.Printf("📤 Sending: [%d] %s", msg.ID, msg.Title)

	content := buildMessageContent(msg)

	_, err := session.ChannelMessageSend(msg.ChannelID, content)
	if err != nil {
		log.Printf("❌ Failed to send [%d]: %v", msg.ID, err)
		return
	}

	if err := msg.Delete(db); err != nil {
		log.Printf("⚠️ Sent [%d] but failed to mark as sent: %v", msg.ID, err)
		return
	}

	log.Printf("✅ Successfully sent [%d] %s", msg.ID, msg.Title)
}

// buildMessageContent constructs the Discord message
func buildMessageContent(msg *models.ScheduledMessage) string {
	content := msg.Message
	return content
}
