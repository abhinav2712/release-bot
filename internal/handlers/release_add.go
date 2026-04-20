package handlers

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"release-bot/internal/discord"
	"release-bot/internal/models"
	"release-bot/internal/status"
	"release-bot/internal/summary"
)

func (h *Handler) openReleaseAddStatusSelector(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the status for your new branch entry:",
			Flags:   discord.EphemeralFlag,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_add_status_select",
							Placeholder: "Choose status",
							Options:     status.BuildStatusOptions(),
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-add status selector:", err)
	}
}

func (h *Handler) openReleaseAddModal(s *discordgo.Session, i *discordgo.InteractionCreate, selectedStatus string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Add Release Entry",
			CustomID: "release_add_modal|" + selectedStatus,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{CustomID: "branch", Label: "Branch Name", Style: discordgo.TextInputShort, Placeholder: "feature/release-bot", Required: true},
				}},
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{CustomID: "title", Label: "Feature Title", Style: discordgo.TextInputShort, Placeholder: "Release bot modal support", Required: true},
				}},
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{CustomID: "pr", Label: "PR Link", Style: discordgo.TextInputShort, Placeholder: "https://github.com/...", Required: false},
				}},
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{CustomID: "blocker", Label: "Blocker (optional)", Style: discordgo.TextInputParagraph, Placeholder: "Waiting for review from backend team", Required: false},
				}},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-add modal:", err)
	}
}

func (h *Handler) handleReleaseAddModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	parts := strings.SplitN(i.ModalSubmitData().CustomID, "|", 2)
	if len(parts) != 2 {
		discord.Message(s, i, "❌ Could not determine selected status.", true)
		return
	}

	data := i.ModalSubmitData().Components
	item := models.ReleaseItem{
		DeveloperID:   i.Member.User.ID,
		DeveloperName: i.Member.User.Username,
		Branch:        strings.TrimSpace(discord.GetModalValue(data, "branch")),
		Title:         strings.TrimSpace(discord.GetModalValue(data, "title")),
		Status:        parts[1],
		PRLink:        strings.TrimSpace(discord.GetModalValue(data, "pr")),
		Blocker:       strings.TrimSpace(discord.GetModalValue(data, "blocker")),
	}

	h.store.AppendItem(item)
	r := h.store.Get()
	summary.Update(s, &r)
	discord.Message(s, i, "✅ Entry added successfully.", true)
}
