package main

import (
	"log"
	"os"

	bot "github.com/betauia/BetaBot.go/bot"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file using the godotenv package
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the database
	bot.InitDatabase()

	// Retrieve the bot token from the environment variable
	bot.BotToken = os.Getenv("BOT_TOKEN")
	if bot.BotToken == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}

	// Run the bot
	bot.Run()
}
