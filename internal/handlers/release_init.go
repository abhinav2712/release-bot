package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"release-bot/internal/discord"
	"release-bot/internal/models"
	"release-bot/internal/status"
	"release-bot/internal/summary"
)

var caser = cases.Title(language.Und)

func (h *Handler) handleReleaseInit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select release type:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_type_select",
							Placeholder: "Choose release type",
							Options: []discordgo.SelectMenuOption{
								{Label: "Major Release", Value: "major"},
								{Label: "Minor Release", Value: "minor"},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error responding to release-init:", err)
	}
}

func (h *Handler) openReleaseDateSelector(s *discordgo.Session, i *discordgo.InteractionCreate, releaseType string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Select release date:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_date_select|" + releaseType,
							Placeholder: "Choose release date",
							Options:     status.BuildReleaseDateOptions(),
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release date selector:", err)
	}
}

func (h *Handler) openReleaseNotesModal(s *discordgo.Session, i *discordgo.InteractionCreate, releaseType, releaseDate string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Release Items",
			CustomID: "release_init_modal|" + releaseType + "|" + releaseDate,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "release_points",
							Label:    "Release items / bugs / enhancements",
							Style:    discordgo.TextInputParagraph,
							Placeholder: "• Fix login issue - <@123456789012345678>\n" +
								"• Add dashboard enhancement - <@987654321098765432>",
							Required: false,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release notes modal:", err)
	}
}

func (h *Handler) handleReleaseInitModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	parts := strings.SplitN(i.ModalSubmitData().CustomID, "|", 3)
	if len(parts) != 3 {
		discord.Message(s, i, "❌ Could not determine release type or date.", true)
		return
	}
	releaseType := parts[1]
	releaseDate := parts[2]
	releaseNotes := strings.TrimSpace(discord.GetModalValue(i.ModalSubmitData().Components, "release_points"))
	h.createReleaseThread(s, i, releaseType, releaseDate, releaseNotes)
}

func (h *Handler) createReleaseThread(s *discordgo.Session, i *discordgo.InteractionCreate, releaseType, releaseDate, releaseNotes string) {
	channelID := i.ChannelID
	threadName := fmt.Sprintf("%s Release - %s", caser.String(releaseType), status.FormatDate(releaseDate))

	msg, err := s.ChannelMessageSend(channelID, "Initializing release...")
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}

	thread, err := s.MessageThreadStart(channelID, msg.ID, threadName, 60)
	if err != nil {
		log.Println("Error creating thread:", err)
		return
	}

	release := models.CurrentRelease{
		ThreadID:     thread.ID,
		ReleaseType:  releaseType,
		ReleaseDate:  releaseDate,
		ReleaseNotes: releaseNotes,
		Items:        []models.ReleaseItem{},
	}

	summaryMsg, err := s.ChannelMessageSend(thread.ID, summary.BuildContent(&release))
	if err != nil {
		log.Println("Error sending summary:", err)
		return
	}

	release.SummaryMsgID = summaryMsg.ID
	h.store.Set(release)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ %s release created for %s: <#%s>", caser.String(releaseType), status.FormatDate(releaseDate), thread.ID),
			Flags:   discord.EphemeralFlag,
		},
	})
	if err != nil {
		log.Println("Error responding after release creation:", err)
	}

	log.Printf("Thread ID: %s | Summary Message ID: %s", thread.ID, summaryMsg.ID)
}
