package status

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// BuildStatusOptions returns the fixed set of status dropdown options.
func BuildStatusOptions() []discordgo.SelectMenuOption {
	return []discordgo.SelectMenuOption{
		{Label: "In Progress", Value: "in-progress", Description: "Work is ongoing"},
		{Label: "Given for Review", Value: "given-for-review", Description: "Shared for review"},
		{Label: "Reviewed", Value: "reviewed", Description: "Review is done"},
		{Label: "Tested", Value: "tested", Description: "Testing is done"},
		{Label: "Reviewed and Tested", Value: "reviewed-and-tested", Description: "Review and testing are done"},
	}
}

// Emoji maps a status value to its display emoji.
func Emoji(s string) string {
	switch s {
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

// Humanize converts an internal status value to a display label.
func Humanize(s string) string {
	switch s {
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
		return s
	}
}

// BuildReleaseDateOptions returns today + next 7 days as dropdown options (DD-MM-YYYY).
func BuildReleaseDateOptions() []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, 8)
	now := time.Now()
	for dayOffset := 0; dayOffset <= 7; dayOffset++ {
		date := now.AddDate(0, 0, dayOffset)
		value := date.Format("02-01-2006")
		description := date.Format("Monday")
		if dayOffset == 0 {
			description = "Today • " + description
		}
		options = append(options, discordgo.SelectMenuOption{
			Label:       value,
			Value:       value,
			Description: description,
		})
	}
	return options
}

// Truncate shortens text to max runes, appending "..." if trimmed.
func Truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}
