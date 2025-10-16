package bot

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/betauia/BetaBot.go/bot/commands"
	"github.com/betauia/BetaBot.go/bot/scheduler"
	"github.com/betauia/BetaBot.go/bot/utils"
	"github.com/bwmarrin/discordgo"
)

var BotToken string
var RemoveCommands = flag.Bool("remove-command", true, "Remove all commands after shutting down or not")

func init() { flag.Parse() }

func Run() {
	if BotToken == "" {
		log.Fatal("No bot token provided. Please set the BOT_TOKEN environment variable.")
	}

	// Inject the database into the commands package
	commands.SetDatabase(DB)

	// Create a new Discord session using the provided bot token.
	discord, err := discordgo.New("Bot " + BotToken)
	utils.CheckNilErr(err)

	// Add interaction handlers
	discord.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			// Handle slash commands
			handlers := commands.GetCommandHandlers()
			if handler, exists := handlers[interaction.ApplicationCommandData().Name]; exists {
				handler(session, interaction)
			}
		case discordgo.InteractionModalSubmit:
			// Handle modal submissions
			modalHandlers := commands.GetModalHandlers()
			if handler, exists := modalHandlers[interaction.ModalSubmitData().CustomID]; exists {
				handler(session, interaction)
			}
		case discordgo.InteractionMessageComponent:
			// Handle button clicks and other components
			componentHandlers := commands.GetComponentHandlers()
			if handler, exists := componentHandlers[interaction.MessageComponentData().CustomID]; exists {
				handler(session, interaction)
			}
		}
	})

	// open session
	discord.Open()
	utils.CheckNilErr(err)
	defer discord.Close() // close session, after function termination

	// Register commands
	commands.RegisterAllCommands(discord)

	// Start scheduler for scheduled messages
	scheduler.Start(discord, DB, 30*time.Second)

	// Keep bot running until a termination signal is received
	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Remove commands if the flag is set
	if *RemoveCommands {
		commands.RemoveCommands(discord)
	}

	fmt.Println("Shutting down bot.")
}
