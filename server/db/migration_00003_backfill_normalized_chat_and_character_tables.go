package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext(
		"00003_backfill_normalized_chat_and_character_tables.go",
		upBackfillNormalizedChatAndCharacterTables,
		downBackfillNormalizedChatAndCharacterTables,
	)
}

type legacyCharacterRow struct {
	ID         string
	Name       string
	MultiVoice int64
	VoiceJSON  string
	ChatJSON   string
}

type legacyChatSessionRow struct {
	SessionID    string
	CharacterID  string
	MessagesJSON string
}

func upBackfillNormalizedChatAndCharacterTables(ctx context.Context, tx *sql.Tx) error {
	if err := clearNormalizedCharacterAndChatTables(ctx, tx); err != nil {
		return err
	}

	characters, err := listLegacyCharacters(ctx, tx)
	if err != nil {
		return err
	}
	for _, character := range characters {
		if err := backfillLegacyCharacter(ctx, tx, character); err != nil {
			return err
		}
	}

	sessions, err := listLegacyChatSessions(ctx, tx)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if err := backfillLegacyChatSession(ctx, tx, session); err != nil {
			return err
		}
	}

	return nil
}

func downBackfillNormalizedChatAndCharacterTables(ctx context.Context, tx *sql.Tx) error {
	return clearNormalizedCharacterAndChatTables(ctx, tx)
}

func clearNormalizedCharacterAndChatTables(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		"DELETE FROM chat_messages",
		"DELETE FROM character_voice_behaviors",
		"DELETE FROM character_voices",
		"DELETE FROM character_chat_limits",
		"DELETE FROM character_chat_settings",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("clear normalized tables: %w", err)
		}
	}

	return nil
}

func listLegacyCharacters(ctx context.Context, tx *sql.Tx) ([]legacyCharacterRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, multi_voice, voice_json, chat_json
		FROM characters
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query legacy characters: %w", err)
	}
	defer rows.Close()

	var characters []legacyCharacterRow
	for rows.Next() {
		var row legacyCharacterRow
		if err := rows.Scan(&row.ID, &row.Name, &row.MultiVoice, &row.VoiceJSON, &row.ChatJSON); err != nil {
			return nil, fmt.Errorf("scan legacy character: %w", err)
		}
		characters = append(characters, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy characters: %w", err)
	}

	return characters, nil
}

func backfillLegacyCharacter(ctx context.Context, tx *sql.Tx, row legacyCharacterRow) error {
	var voices []data.CharacterConfigVoice
	if err := json.Unmarshal([]byte(row.VoiceJSON), &voices); err != nil {
		return fmt.Errorf("unmarshal legacy voices for character %s: %w", row.ID, err)
	}

	var chat data.CharacterConfigChat
	if err := json.Unmarshal([]byte(row.ChatJSON), &chat); err != nil {
		return fmt.Errorf("unmarshal legacy chat for character %s: %w", row.ID, err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO character_chat_settings (
			character_id,
			type,
			model,
			system_prompt,
			temperature_enable,
			temperature_value,
			max_history
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		row.ID,
		chat.Type,
		chat.Model,
		chat.SystemPrompt,
		boolToInt(chat.Temperature.Enable),
		chat.Temperature.Value,
		chat.MaxHistory,
	); err != nil {
		return fmt.Errorf("insert character_chat_settings for %s: %w", row.ID, err)
	}

	limits := []struct {
		window string
		value  data.CharacterConfigChatLimitType
	}{
		{window: "day", value: chat.Limit.Day},
		{window: "hour", value: chat.Limit.Hour},
		{window: "minute", value: chat.Limit.Minute},
	}
	for _, limit := range limits {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_chat_limits (
				character_id,
				window,
				request_limit,
				token_limit
			) VALUES (?, ?, ?, ?)
		`,
			row.ID,
			limit.window,
			limit.value.Request,
			limit.value.Token,
		); err != nil {
			return fmt.Errorf("insert character_chat_limits for %s: %w", row.ID, err)
		}
	}

	for voiceIndex, voice := range voices {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_voices (
				character_id,
				voice_index,
				name,
				type,
				identification,
				model_id,
				model_file,
				speaker_id,
				reference_audio_path,
				image,
				background_image_path
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			row.ID,
			voiceIndex,
			voice.Name,
			voice.Type,
			voice.Identification,
			voice.ModelID,
			voice.ModelFile,
			voice.SpeakerID,
			voice.ReferenceAudioPath,
			voice.Image,
			voice.BackgroundImagePath,
		); err != nil {
			return fmt.Errorf("insert character_voices for %s[%d]: %w", row.ID, voiceIndex, err)
		}

		for behaviorIndex, behavior := range voice.Behavior {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO character_voice_behaviors (
					character_id,
					voice_index,
					behavior_index,
					identification,
					image_path
				) VALUES (?, ?, ?, ?, ?)
			`,
				row.ID,
				voiceIndex,
				behaviorIndex,
				behavior.Identification,
				behavior.ImagePath,
			); err != nil {
				return fmt.Errorf("insert character_voice_behaviors for %s[%d][%d]: %w", row.ID, voiceIndex, behaviorIndex, err)
			}
		}
	}

	return nil
}

func listLegacyChatSessions(ctx context.Context, tx *sql.Tx) ([]legacyChatSessionRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT session_id, character_id, messages_json
		FROM chat_sessions
		ORDER BY session_id, character_id
	`)
	if err != nil {
		return nil, fmt.Errorf("query legacy chat sessions: %w", err)
	}
	defer rows.Close()

	var sessions []legacyChatSessionRow
	for rows.Next() {
		var row legacyChatSessionRow
		if err := rows.Scan(&row.SessionID, &row.CharacterID, &row.MessagesJSON); err != nil {
			return nil, fmt.Errorf("scan legacy chat session: %w", err)
		}
		sessions = append(sessions, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy chat sessions: %w", err)
	}

	return sessions, nil
}

func backfillLegacyChatSession(ctx context.Context, tx *sql.Tx, row legacyChatSessionRow) error {
	var messages []data.ChatCompletionMessage
	if err := json.Unmarshal([]byte(row.MessagesJSON), &messages); err != nil {
		return fmt.Errorf("unmarshal legacy chat messages for %s/%s: %w", row.SessionID, row.CharacterID, err)
	}

	for messageIndex, message := range messages {
		imageExtension := ""
		var imageData []byte
		if message.Image != nil {
			imageExtension = message.Image.Extension
			imageData = message.Image.Data
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO chat_messages (
				session_id,
				character_id,
				message_index,
				role,
				content,
				image_extension,
				image_data
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
			row.SessionID,
			row.CharacterID,
			messageIndex,
			message.Role,
			message.Content,
			imageExtension,
			imageData,
		); err != nil {
			return fmt.Errorf("insert chat_messages for %s/%s[%d]: %w", row.SessionID, row.CharacterID, messageIndex, err)
		}
	}

	return nil
}
