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
// Items are grouped into Frontend and Backend sections.
// Pure function — no side effects, safe to call from anywhere.
func BuildContent(r *models.CurrentRelease) string {
	content := fmt.Sprintf(
		"## %s Release\nrelease date: %s\n\n",
		caser.String(r.ReleaseType),
		status.FormatDate(r.ReleaseDate),
	)

	if r.ReleaseNotes != "" {
		content += fmt.Sprintf("release items:\n%s\n\n", r.ReleaseNotes)
	}

	if len(r.Items) == 0 {
		content += "_No entries yet_"
		return content
	}

	// Bucket items by layer.
	var frontend, backend, other []models.ReleaseItem
	for _, item := range r.Items {
		switch item.Layer {
		case "frontend":
			frontend = append(frontend, item)
		case "backend":
			backend = append(backend, item)
		default:
			other = append(other, item)
		}
	}

	// Render each group; skip header if the group is empty.
	content += renderGroup("🖥️  Frontend", frontend)
	content += renderGroup("⚙️  Backend", backend)

	// Legacy/ungrouped items (added before the layer field existed) rendered without a header.
	for _, item := range other {
		content += renderItem(item)
	}

	return strings.TrimRight(content, "\n")
}

// renderGroup renders a labelled section for a slice of items.
// Returns an empty string when items is nil/empty.
func renderGroup(header string, items []models.ReleaseItem) string {
	if len(items) == 0 {
		return ""
	}
	s := fmt.Sprintf("**%s**\n", header)
	for _, item := range items {
		s += renderItem(item)
	}
	return s + "\n"
}

// renderItem formats a single ReleaseItem.
// Removed items are rendered with Discord strikethrough; active items show full detail.
func renderItem(item models.ReleaseItem) string {
	if item.Status == "removed" {
		return fmt.Sprintf(
			"- ❌ ~~<@%s> — `%s`~~ *(removed)*\n\n",
			item.DeveloperID,
			item.Branch,
		)
	}

	s := fmt.Sprintf(
		"- %s <@%s> — `%s`\n",
		status.Emoji(item.Status),
		item.DeveloperID,
		item.Branch,
	)
	s += fmt.Sprintf("  status: %s\n", status.Humanize(item.Status))
	if item.Title != "" {
		s += fmt.Sprintf("  title: %s\n", item.Title)
	}
	if item.PRLink != "" {
		s += fmt.Sprintf("  pr: %s\n", item.PRLink)
	}
	if item.Blocker != "" {
		s += fmt.Sprintf("  blocker: %s\n", item.Blocker)
	}
	return s + "\n"
}

// Update rebuilds and edits the pinned summary message in the release thread.
func Update(s *discordgo.Session, r *models.CurrentRelease) {
	_, err := s.ChannelMessageEdit(r.ThreadID, r.SummaryMsgID, BuildContent(r))
	if err != nil {
		log.Println("Error updating summary message:", err)
	}
}
