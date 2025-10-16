package commands

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/betauia/BetaBot.go/bot/models"
	"github.com/bwmarrin/discordgo"
)

// Temporary storage for pending scheduled messages
type PendingSchedule struct {
	ChannelID     string
	Title         string
	Message       string
	ScheduledTime time.Time
}

var (
	pendingSchedules = make(map[string]*PendingSchedule) // key: userID
	pendingMutex     sync.RWMutex
)

// Define the schedule command
var scheduleCommand = &discordgo.ApplicationCommand{
	Name:        "schedule",
	Description: "Manage scheduled messages or add a new one",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List all coming scheduled messages for this server",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add a new scheduled message",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "preview",
			Description: "Preview a scheduled message by its ID",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "remove",
			Description: "Remove a scheduled message by its ID",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "edit",
			Description: "Edit a scheduled message by its ID",
		},
	},
}

// Register function to create the "schedule" command
func RegisterScheduleCommand(session *discordgo.Session) *discordgo.ApplicationCommand {
	return scheduleCommand
}

func handleScheduleCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}

	switch options[0].Name {
	case "add":
		handleScheduleAddCommand(session, interaction)
	case "list":
		handleScheduleListCommand(session, interaction)
	case "preview":
		// TODO
	case "remove":
		// TODO
	case "edit":
		// TODO
	}
}

/*
#------------------------------#
|                              |
|       Command handlers       |
|                              |
#------------------------------#
*/

// Handle the "add" subcommand
func handleScheduleAddCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	// Show channel selector first
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select a channel for the scheduled message:",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "schedule_channel_select",
							Placeholder: "Choose a channel",
							MenuType:    discordgo.ChannelSelectMenu,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
								discordgo.ChannelTypeGuildNews,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Failed to send channel selector: %v", err)
	}
}

// Handle channel selection and show modal
func handleChannelSelect(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.MessageComponentData()
	selectedChannelID := data.Values[0]

	// Store the channel ID temporarily
	userID := interaction.Member.User.ID
	pendingMutex.Lock()
	if pendingSchedules[userID] == nil {
		pendingSchedules[userID] = &PendingSchedule{}
	}
	pendingSchedules[userID].ChannelID = selectedChannelID
	pendingMutex.Unlock()

	// Show modal without channel field
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "schedule_add_modal",
			Title:    "Add Scheduled Message",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "time",
							Label:       "Scheduled Time (Oslo timezone)",
							Style:       discordgo.TextInputShort,
							Placeholder: "e.g., 31.12.2025 16:12 or 28.02.2025",
							Required:    true,
							MaxLength:   30,
							MinLength:   1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "title",
							Label:       "Title of the message (unique)",
							Style:       discordgo.TextInputShort,
							Placeholder: "Weekly Update (01.01.2025)",
							Required:    true,
							MaxLength:   25,
							MinLength:   5,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "message",
							Label:       "Message Content",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Enter your Discord Markdown message here...",
							Required:    true,
							MaxLength:   4000,
							MinLength:   1,
						},
					},
				},
			},
		},
	}

	err := session.InteractionRespond(interaction.Interaction, modal)
	if err != nil {
		log.Printf("Failed to send modal: %v", err)
	}

	// Delete the original channel selector message after modal is shown
	go func() {
		time.Sleep(500 * time.Millisecond) // delay to make sure the modal is displayed
		err := session.InteractionResponseDelete(interaction.Interaction)
		if err != nil {
			log.Printf("Failed to delete channel selector message: %v", err)
		}
	}()
}

// handle modal submit
func handleModalSubmit(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.ModalSubmitData()
	timestr := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	title := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	message := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// Validate time format
	scheduledTime, err := parseScheduledTime(timestr)
	if err != nil {
		respondWithError(session, interaction, "Invalid time format. Use formats like 31.12.2025 16:12 or 28.02.2025")
		return
	}

	// Retrieve the channel from pending schedule
	userID := interaction.Member.User.ID
	pendingMutex.Lock()
	pending := pendingSchedules[userID]
	if pending == nil {
		pendingMutex.Unlock()
		respondWithError(session, interaction, "Session expired. Please try again.")
		return
	}
	channel := pending.ChannelID

	// Update with new data
	pending.Title = title
	pending.Message = message
	pending.ScheduledTime = scheduledTime
	pendingMutex.Unlock()

	// Show a preview of the message
	preview := fmt.Sprintf("**Channel:** <#%s>\n**Time:** %s\n**Title:** %s\n**Message:**\n%s", channel, scheduledTime.Format("02.01.2006 15:04 MST"), title, message)
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: preview,
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							CustomID: "confirm_schedule",
							Label:    "Confirm",
							Style:    discordgo.PrimaryButton,
						},
						discordgo.Button{
							CustomID: "cancel_schedule",
							Label:    "Cancel",
							Style:    discordgo.SecondaryButton,
						},
					},
				},
			},
		},
	})
}

func handleScheduleConfirmation(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.MessageComponentData()
	userID := interaction.Member.User.ID

	switch data.CustomID {
	case "confirm_schedule":
		// Retrieve the pending schedule from memory
		pendingMutex.RLock()
		pending, exists := pendingSchedules[userID]
		pendingMutex.RUnlock()

		if !exists {
			respondWithError(session, interaction, "Session expired. Please try again.")
			return
		}

		// Create the scheduled message model
		scheduledMsg := &models.ScheduledMessage{
			Title:         pending.Title,
			GuildID:       interaction.GuildID,
			RoleID:        "",
			UserID:        userID,
			Message:       pending.Message,
			ScheduledTime: pending.ScheduledTime,
			ChannelID:     pending.ChannelID,
		}

		// Save to database
		err := scheduledMsg.Create(db)
		if err != nil {
			respondWithError(session, interaction, fmt.Sprintf("Failed to save scheduled message: %v", err))
			return
		}

		// Clean up the pending schedule
		pendingMutex.Lock()
		delete(pendingSchedules, userID)
		pendingMutex.Unlock()

		// Update the message to remove buttons and show success
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    fmt.Sprintf("✅ Scheduled message saved successfully with ID: %d", scheduledMsg.ID),
				Components: []discordgo.MessageComponent{}, // Remove buttons
				Flags:      discordgo.MessageFlagsEphemeral,
			},
		})
	case "cancel_schedule":
		// Clean up the pending schedule
		pendingMutex.Lock()
		delete(pendingSchedules, userID)
		pendingMutex.Unlock()

		// Update the message to remove buttons and show cancellation
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "❌ Schedule creation canceled.",
				Components: []discordgo.MessageComponent{}, // Remove buttons
				Flags:      discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

// Handler for /schedule list - Only shows messages in that guild
func handleScheduleListCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	guildID := interaction.GuildID

	messages, err := models.GetUpcomingMessagesByGuild(db, guildID)
	if err != nil {
		respondWithError(session, interaction, fmt.Sprintf("Error getting messages from database: %v", err))
		return
	}

	if len(messages) == 0 {
		respondWithSuccess(session, interaction, "No upcoming scheduled messages for this server.")
		return
	}

	response := fmt.Sprintf("**Upcoming Scheduled Messages (%d):**\n\n", len(messages))

	for i, msg := range messages {
		response += fmt.Sprintf("**%d. %s** (ID: %d)\n", i+1, msg.Title, msg.ID)
		response += fmt.Sprintf("   📅 Scheduled: %s\n", msg.ScheduledTime.Format("02.01.2006 15:04 MST"))
		response += fmt.Sprintf("   📢 Channel: <#%s>\n\n", msg.ChannelID)
	}

	respondWithSuccess(session, interaction, response)
}

/*
#------------------------------#
|                              |
|      Utility functions       |
|                              |
#------------------------------#
*/
func respondWithSuccess(session *discordgo.Session, interaction *discordgo.InteractionCreate, message string) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral, // message only appears to the user
		},
	})
}

func respondWithError(session *discordgo.Session, interaction *discordgo.InteractionCreate, message string) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral, // message only appears to the user
		},
	})
}

// parseScheduledTime attempts to parse the time string using multiple formats
func parseScheduledTime(timestr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,              // 2025-01-01T23:59:00Z
		"01.02.2006 15:04 MST",    // 12.31.2025 16:12 CET
		"01.02.2006 15:04:05 MST", // 12.31.2025 16:12:00 CET
		"2006-01-02 15:04 MST",    // 2025-12-31 16:12 CET
		"2006-01-02 15:04:05 MST", // 2025-12-31 16:12:00 CET
		"02.01.2006 15:04 MST",    // 28.02.2025 16:12:00 CET (DD.MM.YYYY)
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, timestr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, lastErr
}
