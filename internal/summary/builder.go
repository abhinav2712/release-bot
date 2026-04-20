package summary

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"release-bot/internal/models"
	"release-bot/internal/status"
)

var caser = cases.Title(language.Und)

// BuildContent constructs the full summary message string from the current release state.
// Pure function — no side effects, safe to call from anywhere.
func BuildContent(r *models.CurrentRelease) string {
	content := fmt.Sprintf(
		"## %s Release\nrelease date: %s\n\n",
		caser.String(r.ReleaseType),
		r.ReleaseDate,
	)

	if r.ReleaseNotes != "" {
		content += fmt.Sprintf("release items:\n%s\n\n", r.ReleaseNotes)
	}

	if len(r.Items) == 0 {
		content += "_No entries yet_"
	} else {
		for _, item := range r.Items {
			if item.Status == "removed" {
				// Render removed entries with strikethrough — visually struck out, no detail lines.
				content += fmt.Sprintf(
					"- ❌ ~~<@%s> — `%s`~~ *(removed)*\n\n",
					item.DeveloperID,
					item.Branch,
				)
				continue
			}
			content += fmt.Sprintf(
				"- %s <@%s> — `%s`\n",
				status.Emoji(item.Status),
				item.DeveloperID,
				item.Branch,
			)
			content += fmt.Sprintf("  status: %s\n", status.Humanize(item.Status))
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
		content = strings.TrimRight(content, "\n")
	}

	return content
}

// Update rebuilds and edits the pinned summary message in the release thread.
func Update(s *discordgo.Session, r *models.CurrentRelease) {
	_, err := s.ChannelMessageEdit(r.ThreadID, r.SummaryMsgID, BuildContent(r))
	if err != nil {
		log.Println("Error updating summary message:", err)
	}
}
