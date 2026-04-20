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

// openReleaseAddLayerSelector is the first step of /release-add.
// The user picks Frontend or Backend before choosing a status.
func (h *Handler) openReleaseAddLayerSelector(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Is this a Frontend or Backend branch?",
			Flags:   discord.EphemeralFlag,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_add_layer_select",
							Placeholder: "Choose layer",
							Options: []discordgo.SelectMenuOption{
								{Label: "🖥️ Frontend", Value: "frontend", Description: "UI / client-side branch"},
								{Label: "⚙️ Backend", Value: "backend", Description: "API / server-side branch"},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-add layer selector:", err)
	}
}

// openReleaseAddStatusSelector is the second step — user picks a status after choosing a layer.
func (h *Handler) openReleaseAddStatusSelector(s *discordgo.Session, i *discordgo.InteractionCreate, layer string) {
	layerLabel := layerLabel(layer)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the status for your **" + layerLabel + "** branch entry:",
			Flags:   discord.EphemeralFlag,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_add_status_select|" + layer,
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

func (h *Handler) openReleaseAddModal(s *discordgo.Session, i *discordgo.InteractionCreate, layer, selectedStatus string) {
	layerLabel := layerLabel(layer)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Add " + layerLabel + " Entry",
			CustomID: "release_add_modal|" + layer + "|" + selectedStatus,
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
	// customID format: release_add_modal|<layer>|<status>
	parts := strings.SplitN(i.ModalSubmitData().CustomID, "|", 3)
	if len(parts) != 3 {
		discord.Message(s, i, "❌ Could not determine layer or status.", true)
		return
	}

	layer := parts[1]
	selectedStatus := parts[2]
	data := i.ModalSubmitData().Components

	item := models.ReleaseItem{
		DeveloperID:   i.Member.User.ID,
		DeveloperName: i.Member.User.Username,
		Layer:         layer,
		Branch:        strings.TrimSpace(discord.GetModalValue(data, "branch")),
		Title:         strings.TrimSpace(discord.GetModalValue(data, "title")),
		Status:        selectedStatus,
		PRLink:        strings.TrimSpace(discord.GetModalValue(data, "pr")),
		Blocker:       strings.TrimSpace(discord.GetModalValue(data, "blocker")),
	}

	h.store.AppendItem(item)
	r := h.store.Get()
	summary.Update(s, &r)
	discord.Message(s, i, "✅ "+layerLabel(layer)+" entry added successfully.", true)
}

// layerLabel returns a display-friendly name for the given layer value.
func layerLabel(layer string) string {
	switch layer {
	case "frontend":
		return "Frontend"
	case "backend":
		return "Backend"
	default:
		return "General"
	}
}
