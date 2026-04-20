package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// EphemeralFlag is the Discord message flag that makes a message only visible to the invoking user.
const EphemeralFlag = discordgo.MessageFlagsEphemeral

// Message sends a plain-text interaction response, optionally ephemeral.
func Message(s *discordgo.Session, i *discordgo.InteractionCreate, message string, ephemeral bool) {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = EphemeralFlag
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   flags,
		},
	})
	if err != nil {
		log.Println("Error sending interaction response:", err)
	}
}

// UpdateMessage edits an existing interaction message (used for select-menu step transitions).
func UpdateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string, components []discordgo.MessageComponent) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    message,
			Components: components,
		},
	})
	if err != nil {
		log.Println("Error updating interaction message:", err)
	}
}

// GetModalValue extracts the value of a TextInput field from modal submit data by its customID.
// This was previously duplicated three times in main.go as a local closure.
func GetModalValue(data []discordgo.MessageComponent, id string) string {
	for _, row := range data {
		actionRow, ok := row.(*discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, comp := range actionRow.Components {
			textInput, ok := comp.(*discordgo.TextInput)
			if !ok {
				continue
			}
			if textInput.CustomID == id {
				return textInput.Value
			}
		}
	}
	return ""
}
