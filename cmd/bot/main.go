package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type ReleaseItem struct {
	DeveloperID   string
	DeveloperName string
	Branch        string
	Title         string
	Status        string
	PRLink        string
	Blocker       string
}

type CurrentRelease struct {
	ThreadID     string
	SummaryMsgID string
	ReleaseType  string
	ReleaseDate  string
	ReleaseNotes string
	Items        []ReleaseItem
}

var currentRelease CurrentRelease

const ephemeralFlag = discordgo.MessageFlagsEphemeral

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is empty")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	dg.AddHandler(ready)
	dg.AddHandler(interactionHandler)

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}

	fmt.Println("Bot is running...")

	registerCommands(dg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Bot is ready!")
}

func registerCommands(s *discordgo.Session) {
	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		log.Fatal("GUILD_ID is empty")
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Ping test",
	})
	if err != nil {
		log.Fatal("Cannot create ping command:", err)
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        "release-init",
		Description: "Initialize a new release",
	})
	if err != nil {
		log.Fatal("Cannot create release-init command:", err)
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        "release-add",
		Description: "Add your branch to the current release",
	})
	if err != nil {
		log.Fatal("Cannot create release-add command:", err)
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        "release-update",
		Description: "Update one of your branches in the current release",
	})
	if err != nil {
		log.Fatal("Cannot create release-update command:", err)
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        "release-summary",
		Description: "Show quick summary of current release",
	})
	if err != nil {
		log.Fatal("Cannot create release-summary command:", err)
	}
}

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		handleApplicationCommand(s, i)
	case discordgo.InteractionMessageComponent:
		handleMessageComponent(s, i)
	case discordgo.InteractionModalSubmit:
		handleModalSubmit(s, i)
	}
}

func handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "ping":
		respondMessage(s, i, "Pong 🚀", false)

	case "release-init":
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

	case "release-add":
		if currentRelease.ThreadID == "" {
			respondMessage(s, i, "No active release found. Run /release-init first.", true)
			return
		}
		openReleaseAddStatusSelector(s, i)

	case "release-update":
		if currentRelease.ThreadID == "" {
			respondMessage(s, i, "No active release found. Run /release-init first.", true)
			return
		}
		openReleaseUpdateBranchSelector(s, i)

	case "release-summary":
		if currentRelease.ThreadID == "" {
			respondMessage(s, i, "No active release found. Run /release-init first.", true)
			return
		}
		handleReleaseSummary(s, i)
	}
}

func handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	switch {
	case customID == "release_type_select":
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			respondUpdateMessage(s, i, "No release type selected.", nil)
			return
		}
		releaseType := values[0]
		openReleaseDateSelector(s, i, releaseType)

	case strings.HasPrefix(customID, "release_date_select|"):
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			respondUpdateMessage(s, i, "No release date selected.", nil)
			return
		}

		parts := strings.SplitN(customID, "|", 2)
		if len(parts) != 2 {
			respondUpdateMessage(s, i, "Could not determine release type.", nil)
			return
		}

		releaseType := parts[1]
		releaseDate := values[0]
		openReleaseNotesModal(s, i, releaseType, releaseDate)

	case customID == "release_add_status_select":
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			respondUpdateMessage(s, i, "No status selected.", nil)
			return
		}
		selectedStatus := values[0]
		openReleaseAddModal(s, i, selectedStatus)

	case customID == "release_update_branch_select":
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			respondUpdateMessage(s, i, "No branch selected.", nil)
			return
		}
		selectedBranch := values[0]
		openReleaseUpdateStatusSelector(s, i, selectedBranch)

	case strings.HasPrefix(customID, "release_update_status_select|"):
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			respondUpdateMessage(s, i, "No status selected.", nil)
			return
		}

		parts := strings.SplitN(customID, "|", 2)
		if len(parts) != 2 {
			respondUpdateMessage(s, i, "Could not determine branch to update.", nil)
			return
		}

		selectedBranch := parts[1]
		selectedStatus := values[0]
		openReleaseUpdateModal(s, i, selectedBranch, selectedStatus)
	}
}

func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch {
	case strings.HasPrefix(i.ModalSubmitData().CustomID, "release_init_modal|"):
		handleReleaseInitModal(s, i)

	case strings.HasPrefix(i.ModalSubmitData().CustomID, "release_add_modal|"):
		handleReleaseAdd(s, i)

	case strings.HasPrefix(i.ModalSubmitData().CustomID, "release_update_modal|"):
		handleReleaseUpdate(s, i)
	}
}

func openReleaseDateSelector(s *discordgo.Session, i *discordgo.InteractionCreate, releaseType string) {
	customID := "release_date_select|" + releaseType

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Select release date:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    customID,
							Placeholder: "Choose release date",
							Options:     buildReleaseDateOptions(),
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

func buildReleaseDateOptions() []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, 8)
	now := time.Now()

	for dayOffset := 0; dayOffset <= 7; dayOffset++ {
		date := now.AddDate(0, 0, dayOffset)
		value := date.Format("02-01-2006")
		label := date.Format("02-01-2006")
		description := date.Format("Monday")

		if dayOffset == 0 {
			description = "Today • " + description
		}

		options = append(options, discordgo.SelectMenuOption{
			Label:       label,
			Value:       value,
			Description: description,
		})
	}

	return options
}

func openReleaseNotesModal(s *discordgo.Session, i *discordgo.InteractionCreate, releaseType string, releaseDate string) {
	customID := "release_init_modal|" + releaseType + "|" + releaseDate

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Release Items",
			CustomID: customID,
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

func handleReleaseInitModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.ModalSubmitData().CustomID
	parts := strings.SplitN(customID, "|", 3)
	if len(parts) != 3 {
		respondMessage(s, i, "❌ Could not determine release type or date.", true)
		return
	}

	releaseType := parts[1]
	releaseDate := parts[2]

	data := i.ModalSubmitData().Components

	getValue := func(customID string) string {
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

				if textInput.CustomID == customID {
					return textInput.Value
				}
			}
		}
		return ""
	}

	releaseNotes := strings.TrimSpace(getValue("release_points"))
	createReleaseThread(s, i, releaseType, releaseDate, releaseNotes)
}

func createReleaseThread(s *discordgo.Session, i *discordgo.InteractionCreate, releaseType string, releaseDate string, releaseNotes string) {
	channelID := i.ChannelID
	threadName := fmt.Sprintf("%s-release-%s", releaseType, releaseDate)

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

	summary := fmt.Sprintf("## %s Release\nrelease date: %s\n\n", strings.Title(releaseType), releaseDate)

	if releaseNotes != "" {
		summary += fmt.Sprintf("release items:\n%s\n\n", releaseNotes)
	}

	summary += "_No entries yet_"

	summaryMsg, err := s.ChannelMessageSend(thread.ID, summary)
	if err != nil {
		log.Println("Error sending summary:", err)
		return
	}

	currentRelease.ThreadID = thread.ID
	currentRelease.SummaryMsgID = summaryMsg.ID
	currentRelease.ReleaseType = releaseType
	currentRelease.ReleaseDate = releaseDate
	currentRelease.ReleaseNotes = releaseNotes
	currentRelease.Items = []ReleaseItem{}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ %s release created for %s: <#%s>", strings.Title(releaseType), releaseDate, thread.ID),
			Flags:   ephemeralFlag,
		},
	})
	if err != nil {
		log.Println("Error responding after release creation:", err)
	}

	fmt.Println("Thread ID:", thread.ID)
	fmt.Println("Summary Message ID:", summaryMsg.ID)
}

func openReleaseAddStatusSelector(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the status for your new branch entry:",
			Flags:   ephemeralFlag,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "release_add_status_select",
							Placeholder: "Choose status",
							Options:     buildStatusOptions(),
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

func openReleaseAddModal(s *discordgo.Session, i *discordgo.InteractionCreate, status string) {
	customID := "release_add_modal|" + status

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Add Release Entry",
			CustomID: customID,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "branch",
							Label:       "Branch Name",
							Style:       discordgo.TextInputShort,
							Placeholder: "feature/release-bot",
							Required:    true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "title",
							Label:       "Feature Title",
							Style:       discordgo.TextInputShort,
							Placeholder: "Release bot modal support",
							Required:    true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "pr",
							Label:       "PR Link",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://github.com/...",
							Required:    false,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "blocker",
							Label:       "Blocker (optional)",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Waiting for review from backend team",
							Required:    false,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-add modal:", err)
	}
}

func openReleaseUpdateBranchSelector(s *discordgo.Session, i *discordgo.InteractionCreate) {
	developerID := i.Member.User.ID
	options := make([]discordgo.SelectMenuOption, 0)

	for _, item := range currentRelease.Items {
		if item.DeveloperID == developerID {
			description := humanizeStatus(item.Status)
			if item.Title != "" {
				description = fmt.Sprintf("%s • %s", humanizeStatus(item.Status), item.Title)
			}

			options = append(options, discordgo.SelectMenuOption{
				Label:       item.Branch,
				Value:       item.Branch,
				Description: truncate(description, 100),
			})
		}
	}

	if len(options) == 0 {
		respondMessage(s, i, "You do not have any branches in the current release yet.", true)
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the branch you want to update:",
			Flags:   ephemeralFlag,
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

func openReleaseUpdateStatusSelector(s *discordgo.Session, i *discordgo.InteractionCreate, branch string) {
	customID := "release_update_status_select|" + branch

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the new status:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    customID,
							Placeholder: "Choose new status",
							Options:     buildStatusOptions(),
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

func openReleaseUpdateModal(s *discordgo.Session, i *discordgo.InteractionCreate, branch string, status string) {
	customID := "release_update_modal|" + branch + "|" + status

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Update Release Entry",
			CustomID: customID,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "pr",
							Label:       "New PR Link (optional)",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://github.com/...",
							Required:    false,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "blocker",
							Label:       "New Blocker (optional)",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Leave empty if no blocker",
							Required:    false,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error opening release-update modal:", err)
	}
}

func handleReleaseAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData().Components
	customID := i.ModalSubmitData().CustomID

	parts := strings.SplitN(customID, "|", 2)
	if len(parts) != 2 {
		respondMessage(s, i, "❌ Could not determine selected status.", true)
		return
	}

	status := parts[1]

	getValue := func(customID string) string {
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

				if textInput.CustomID == customID {
					return textInput.Value
				}
			}
		}
		return ""
	}

	item := ReleaseItem{
		DeveloperID:   i.Member.User.ID,
		DeveloperName: i.Member.User.Username,
		Branch:        strings.TrimSpace(getValue("branch")),
		Title:         strings.TrimSpace(getValue("title")),
		Status:        status,
		PRLink:        strings.TrimSpace(getValue("pr")),
		Blocker:       strings.TrimSpace(getValue("blocker")),
	}

	currentRelease.Items = append(currentRelease.Items, item)

	updateSummaryMessage(s)
	respondMessage(s, i, "✅ Entry added successfully.", true)
}

func handleReleaseUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData().Components
	customID := i.ModalSubmitData().CustomID

	parts := strings.SplitN(customID, "|", 3)
	if len(parts) != 3 {
		respondMessage(s, i, "❌ Could not determine branch or status to update.", true)
		return
	}

	branch := parts[1]
	newStatus := parts[2]
	developerID := i.Member.User.ID

	getValue := func(customID string) string {
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

				if textInput.CustomID == customID {
					return textInput.Value
				}
			}
		}
		return ""
	}

	newPR := strings.TrimSpace(getValue("pr"))
	newBlocker := strings.TrimSpace(getValue("blocker"))

	found := false

	for index := range currentRelease.Items {
		item := &currentRelease.Items[index]

		if item.DeveloperID == developerID && item.Branch == branch {
			item.Status = newStatus
			item.DeveloperName = i.Member.User.Username

			if newPR != "" {
				item.PRLink = newPR
			}
			item.Blocker = newBlocker
			found = true
			break
		}
	}

	if !found {
		respondMessage(s, i, "❌ No matching branch found for you in the current release.", true)
		return
	}

	updateSummaryMessage(s)
	respondMessage(s, i, fmt.Sprintf("✅ Updated `%s` successfully.", branch), true)
}

func handleReleaseSummary(s *discordgo.Session, i *discordgo.InteractionCreate) {
	counts := map[string]int{
		"in-progress":         0,
		"given-for-review":    0,
		"reviewed":            0,
		"tested":              0,
		"reviewed-and-tested": 0,
	}

	for _, item := range currentRelease.Items {
		counts[item.Status]++
	}

	content := fmt.Sprintf(
		"## %s Release Summary\n"+
			"Release date: **%s**\n\n"+
			"Total branches: **%d**\n"+
			"%s In Progress: **%d**\n"+
			"%s Given for Review: **%d**\n"+
			"%s Reviewed: **%d**\n"+
			"%s Tested: **%d**\n"+
			"%s Reviewed and Tested: **%d**",
		strings.Title(currentRelease.ReleaseType),
		currentRelease.ReleaseDate,
		len(currentRelease.Items),
		statusEmoji("in-progress"), counts["in-progress"],
		statusEmoji("given-for-review"), counts["given-for-review"],
		statusEmoji("reviewed"), counts["reviewed"],
		statusEmoji("tested"), counts["tested"],
		statusEmoji("reviewed-and-tested"), counts["reviewed-and-tested"],
	)

	respondMessage(s, i, content, true)
}

func updateSummaryMessage(s *discordgo.Session) {
	content := fmt.Sprintf(
		"## %s Release\nrelease date: %s\n\n",
		strings.Title(currentRelease.ReleaseType),
		currentRelease.ReleaseDate,
	)

	if currentRelease.ReleaseNotes != "" {
		content += fmt.Sprintf("release items:\n%s\n\n", currentRelease.ReleaseNotes)
	}

	if len(currentRelease.Items) == 0 {
		content += "_No entries yet_"
	} else {
		for _, item := range currentRelease.Items {
			content += fmt.Sprintf(
				"- %s <@%s> — `%s`\n",
				statusEmoji(item.Status),
				item.DeveloperID,
				item.Branch,
			)

			content += fmt.Sprintf("  status: %s\n", humanizeStatus(item.Status))

			if item.Title != "" {
				content += fmt.Sprintf("  title: %s\n", item.Title)
			}
			if item.PRLink != "" {
				content += fmt.Sprintf("  pr: %s\n", item.PRLink)
			}
			if item.Blocker != "" {
				content += fmt.Sprintf("  blocker: %s\n", item.Blocker)
			}

			content += "\n"
		}
	}

	_, err := s.ChannelMessageEdit(currentRelease.ThreadID, currentRelease.SummaryMsgID, content)
	if err != nil {
		log.Println("Error updating summary message:", err)
	}
}

func respondMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string, ephemeral bool) {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = ephemeralFlag
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

func respondUpdateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string, components []discordgo.MessageComponent) {
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

func buildStatusOptions() []discordgo.SelectMenuOption {
	return []discordgo.SelectMenuOption{
		{Label: "In Progress", Value: "in-progress", Description: "Work is ongoing"},
		{Label: "Given for Review", Value: "given-for-review", Description: "Shared for review"},
		{Label: "Reviewed", Value: "reviewed", Description: "Review is done"},
		{Label: "Tested", Value: "tested", Description: "Testing is done"},
		{Label: "Reviewed and Tested", Value: "reviewed-and-tested", Description: "Review and testing are done"},
	}
}

func statusEmoji(status string) string {
	switch status {
	case "in-progress":
		return "🛠️"
	case "given-for-review":
		return "👀"
	case "reviewed":
		return "✅"
	case "tested":
		return "🧪"
	case "reviewed-and-tested":
		return "🚀"
	default:
		return "•"
	}
}

func humanizeStatus(status string) string {
	switch status {
	case "in-progress":
		return "In Progress"
	case "given-for-review":
		return "Given for Review"
	case "reviewed":
		return "Reviewed"
	case "tested":
		return "Tested"
	case "reviewed-and-tested":
		return "Reviewed and Tested"
	default:
		return status
	}
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}