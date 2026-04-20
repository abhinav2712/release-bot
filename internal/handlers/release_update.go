package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"release-bot/internal/discord"
	"release-bot/internal/status"
	"release-bot/internal/summary"
)

func (h *Handler) openReleaseUpdateBranchSelector(s *discordgo.Session, i *discordgo.InteractionCreate) {
	developerID := i.Member.User.ID
	r := h.store.Get()

	var options []discordgo.SelectMenuOption
	for _, item := range r.Items {
		if item.DeveloperID != developerID {
			continue
		}
		description := status.Humanize(item.Status)
		if item.Title != "" {
			description = fmt.Sprintf("%s • %s", status.Humanize(item.Status), item.Title)
		}
		options = append(options, discordgo.SelectMenuOption{
			Label:       item.Branch,
			Value:       item.Branch,
			Description: status.Truncate(description, 100),
		})
	}

	if len(options) == 0 {
		discord.Message(s, i, "You do not have any branches in the current release yet.", true)
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the branch you want to update:",
			Flags:   discord.EphemeralFlag,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_update_branch_select",
							Placeholder: "Choose one of your branches",
							Options:     options,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-update branch selector:", err)
	}
}

func (h *Handler) openReleaseUpdateStatusSelector(s *discordgo.Session, i *discordgo.InteractionCreate, branch string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the new status:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_update_status_select|" + branch,
							Placeholder: "Choose new status",
							Options:     status.BuildStatusOptions(),
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-update status selector:", err)
	}
}

func (h *Handler) openReleaseUpdateModal(s *discordgo.Session, i *discordgo.InteractionCreate, branch, selectedStatus string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Update Release Entry",
			CustomID: "release_update_modal|" + branch + "|" + selectedStatus,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{CustomID: "pr", Label: "New PR Link (optional)", Style: discordgo.TextInputShort, Placeholder: "https://github.com/...", Required: false},
				}},
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{CustomID: "blocker", Label: "New Blocker (optional)", Style: discordgo.TextInputParagraph, Placeholder: "Leave empty if no blocker", Required: false},
				}},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-update modal:", err)
	}
}

func (h *Handler) handleReleaseUpdateModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	parts := strings.SplitN(i.ModalSubmitData().CustomID, "|", 3)
	if len(parts) != 3 {
		discord.Message(s, i, "❌ Could not determine branch or status to update.", true)
		return
	}

	branch := parts[1]
	newStatus := parts[2]
	data := i.ModalSubmitData().Components
	newPR := strings.TrimSpace(discord.GetModalValue(data, "pr"))
	newBlocker := strings.TrimSpace(discord.GetModalValue(data, "blocker"))

	found := h.store.UpdateItem(i.Member.User.ID, i.Member.User.Username, branch, newStatus, newPR, newBlocker)
	if !found {
		discord.Message(s, i, "❌ No matching branch found for you in the current release.", true)
		return
	}

	r := h.store.Get()
	summary.Update(s, &r)
	discord.Message(s, i, fmt.Sprintf("✅ Updated `%s` successfully.", branch), true)
}
