package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"release-bot/internal/discord"
	"release-bot/internal/status"
)

func (h *Handler) handleReleaseSummary(s *discordgo.Session, i *discordgo.InteractionCreate) {
	r := h.store.Get()

	counts := map[string]int{
		"in-progress":          0,
		"given-for-review":     0,
		"reviewed":             0,
		"tested":               0,
		"reviewed-and-tested":  0,
	}
	for _, item := range r.Items {
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
		caser.String(r.ReleaseType),
		r.ReleaseDate,
		len(r.Items),
		status.Emoji("in-progress"), counts["in-progress"],
		status.Emoji("given-for-review"), counts["given-for-review"],
		status.Emoji("reviewed"), counts["reviewed"],
		status.Emoji("tested"), counts["tested"],
		status.Emoji("reviewed-and-tested"), counts["reviewed-and-tested"],
	)

	discord.Message(s, i, content, true)
}
