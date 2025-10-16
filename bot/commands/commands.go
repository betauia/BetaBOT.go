package commands

import (
	"database/sql"
	"fmt"

	"github.com/betauia/BetaBot.go/bot/utils"
	"github.com/bwmarrin/discordgo"
)

var db *sql.DB

func SetDatabase(database *sql.DB) {
	db = database
}

var (
	defaultMemberPermissions int64 = discordgo.PermissionManageGuild

	commands = []*discordgo.ApplicationCommand{
		{
			Name:                     "ping",
			Description:              "Replies with Pong!",
			DefaultMemberPermissions: &defaultMemberPermissions,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Version:                  "0.1.0",
			Type:                     1,
		},
		{
			Name:                     "add",
			Description:              "Adds two numbers.",
			DefaultMemberPermissions: &defaultMemberPermissions,
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "num1",
					Description: "The first number to add",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "num2",
					Description: "The second number to add",
					Required:    true,
				},
			},
			Version: "0.1.0",
			Type:    1,
		},
		scheduleCommand, // Add scheduleCommand
	}

	// Command Handlers - triggered by /commands
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ping":     handlePingCommand,
		"add":      handleAddCommand,
		"schedule": handleScheduleCommand,
	}

	// Modal handlers - triggered when modals are submitted
	modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"schedule_add_modal": handleModalSubmit,
	}

	// Component handlers - triggered when buttons/select menus are clicked
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"schedule_channel_select": handleChannelSelect,
		"confirm_schedule":        handleScheduleConfirmation,
		"cancel_schedule":         handleScheduleConfirmation,
	}
)

func RegisterAllCommands(session *discordgo.Session) {
	// Register the commands
	for _, command := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", command)
		utils.CheckNilErr(err)
	}
}

func RemoveCommands(session *discordgo.Session) {
	for _, cmd := range commands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", cmd.ID)
		utils.CheckNilErr(err)
	}
}

func GetCommandHandlers() map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	return commandHandlers
}

func GetModalHandlers() map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	return modalHandlers
}

func GetComponentHandlers() map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	return componentHandlers
}

// Command Handlers
// Handler for the "ping" command
func handlePingCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	// Respond to the interaction with "Pong!"
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
	utils.CheckNilErr(err)
}

// Handler for the "add" command
func handleAddCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	if len(options) < 2 {
		// Respond with an error message if not enough options are provided
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide two numbers to add.",
			},
		})
		utils.CheckNilErr(err)
		return
	}

	// Extract the numbers from the options
	num1 := options[0].IntValue()
	num2 := options[1].IntValue()

	// Calculate the sum
	sum := num1 + num2

	// Respond with the result
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("The sum of %d and %d is %d.", num1, num2, sum),
		},
	})
	utils.CheckNilErr(err)
}
