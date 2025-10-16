package models

import (
	"database/sql"
	"time"
)

// ScheduledMessage model for a scheduled Discord message
type ScheduledMessage struct {
	ID            int64
	Title         string
	GuildID       string
	RoleID        string
	UserID        string
	Message       string
	ScheduledTime time.Time
	ChannelID     string
}

// Create inserts a new scheduled message into the database
func (sm *ScheduledMessage) Create(db *sql.DB) error {
	query := `
		INSERT INTO scheduled_messages (title, guild_id, role_id, user_id, message, scheduled_time, channel_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query, sm.Title, sm.GuildID, sm.RoleID, sm.UserID, sm.Message, sm.ScheduledTime, sm.ChannelID)
	if err != nil {
		return err
	}

	sm.ID, err = result.LastInsertId()
	return err
}

// GetByID retrieves a scheduled message by ID
func GetScheduledMessageByID(db *sql.DB, id int64) (*ScheduledMessage, error) {
	query := `SELECT id, title, guild_id, role_id, user_id, message, scheduled_time, channel_id FROM scheduled_messages WHERE id = ?`

	sm := &ScheduledMessage{}
	err := db.QueryRow(query, id).Scan(&sm.ID, &sm.Title, &sm.GuildID, &sm.RoleID, &sm.UserID, &sm.Message, &sm.ScheduledTime)
	if err != nil {
		return nil, err
	}

	return sm, nil
}

// GetAllByGuild retrieves all scheduled messages for a guild
func GetScheduledMessagesByGuild(db *sql.DB, guildID string) ([]*ScheduledMessage, error) {
	query := `SELECT id, title, guild_id, role_id, user_id, message, scheduled_time, channel_id FROM scheduled_messages WHERE guild_id = ? ORDER BY scheduled_time ASC`

	rows, err := db.Query(query, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ScheduledMessage
	for rows.Next() {
		sm := &ScheduledMessage{}
		err := rows.Scan(&sm.ID, &sm.Title, &sm.GuildID, &sm.RoleID, &sm.UserID, &sm.Message, &sm.ScheduledTime)
		if err != nil {
			return nil, err
		}
		messages = append(messages, sm)
	}

	return messages, rows.Err()
}

// Update modifies an existing scheduled message
func (sm *ScheduledMessage) Update(db *sql.DB) error {
	query := `
        UPDATE scheduled_messages 
        SET title = ?, guild_id = ?, role_id = ?, user_id = ?, message = ?, scheduled_time = ?, channel_id = ?
        WHERE id = ?
    `
	_, err := db.Exec(query, sm.Title, sm.GuildID, sm.RoleID, sm.UserID, sm.Message, sm.ScheduledTime, sm.ChannelID, sm.ID)
	return err
}

// Delete removes a scheduled message from the database
func (sm *ScheduledMessage) Delete(db *sql.DB) error {
	query := `DELETE FROM scheduled_messages WHERE id = ?`
	_, err := db.Exec(query, sm.ID)
	return err
}

// GetUpcomingMessagesByGuild retrieves all messages scheduled for the future in a guild
func GetUpcomingMessagesByGuild(db *sql.DB, guildID string) ([]*ScheduledMessage, error) {
	query := `SELECT id, title, guild_id, role_id, user_id, message, scheduled_time, channel_id 
        FROM scheduled_messages 
        WHERE guild_id = ? AND scheduled_time > ? 
        ORDER BY scheduled_time ASC`

	rows, err := db.Query(query, guildID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ScheduledMessage
	for rows.Next() {
		sm := &ScheduledMessage{}
		err := rows.Scan(&sm.ID, &sm.Title, &sm.GuildID, &sm.RoleID, &sm.UserID, &sm.Message, &sm.ScheduledTime, &sm.ChannelID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, sm)
	}

	return messages, rows.Err()
}

// GetPendingMessages retrieves all messages that are ready to be sent
func GetPendingMessages(db *sql.DB) ([]*ScheduledMessage, error) {
	query := `SELECT id, title, guild_id, role_id, user_id, message, scheduled_time, channel_id 
        FROM scheduled_messages 
        WHERE scheduled_time <= ? 
        ORDER BY scheduled_time ASC`

	rows, err := db.Query(query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ScheduledMessage
	for rows.Next() {
		sm := &ScheduledMessage{}
		err := rows.Scan(&sm.ID, &sm.Title, &sm.GuildID, &sm.RoleID, &sm.UserID, &sm.Message, &sm.ScheduledTime, &sm.ChannelID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, sm)
	}

	return messages, rows.Err()
}
