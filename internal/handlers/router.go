package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"release-bot/internal/discord"
	"release-bot/internal/state"
)

// Handler holds shared dependencies wired in from main.
// All interaction handler methods are defined on this type across the handlers package files.
type Handler struct {
	store *state.Store
}

// New creates a Handler backed by the given state store.
func New(store *state.Store) *Handler {
	return &Handler{store: store}
}

// InteractionHandler is the top-level Discord interaction dispatcher registered via dg.AddHandler.
func (h *Handler) InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleApplicationCommand(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleMessageComponent(s, i)
	case discordgo.InteractionModalSubmit:
		h.handleModalSubmit(s, i)
	}
}

func (h *Handler) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "ping":
		discord.Message(s, i, "Pong 🚀", false)
	case "release-init":
		h.handleReleaseInit(s, i)
	case "release-add":
		if !h.store.IsActive() {
			discord.Message(s, i, "No active release found. Run /release-init first.", true)
			return
		}
		h.openReleaseAddStatusSelector(s, i)
	case "release-update":
		if !h.store.IsActive() {
			discord.Message(s, i, "No active release found. Run /release-init first.", true)
			return
		}
		h.openReleaseUpdateBranchSelector(s, i)
	case "release-summary":
		if !h.store.IsActive() {
			discord.Message(s, i, "No active release found. Run /release-init first.", true)
			return
		}
		h.handleReleaseSummary(s, i)
	}
}

func (h *Handler) handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	switch {
	case customID == "release_type_select":
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			discord.UpdateMessage(s, i, "No release type selected.", nil)
			return
		}
		h.openReleaseDateSelector(s, i, values[0])

	case strings.HasPrefix(customID, "release_date_select|"):
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			discord.UpdateMessage(s, i, "No release date selected.", nil)
			return
		}
		parts := strings.SplitN(customID, "|", 2)
		if len(parts) != 2 {
			discord.UpdateMessage(s, i, "Could not determine release type.", nil)
			return
		}
		h.openReleaseNotesModal(s, i, parts[1], values[0])

	case customID == "release_add_status_select":
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			discord.UpdateMessage(s, i, "No status selected.", nil)
			return
		}
		h.openReleaseAddModal(s, i, values[0])

	case customID == "release_update_branch_select":
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			discord.UpdateMessage(s, i, "No branch selected.", nil)
			return
		}
		h.openReleaseUpdateStatusSelector(s, i, values[0])

	case strings.HasPrefix(customID, "release_update_status_select|"):
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			discord.UpdateMessage(s, i, "No status selected.", nil)
			return
		}
		parts := strings.SplitN(customID, "|", 2)
		if len(parts) != 2 {
			discord.UpdateMessage(s, i, "Could not determine branch to update.", nil)
			return
		}
		h.openReleaseUpdateModal(s, i, parts[1], values[0])
	}
}

func (h *Handler) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.ModalSubmitData().CustomID
	switch {
	case strings.HasPrefix(customID, "release_init_modal|"):
		h.handleReleaseInitModal(s, i)
	case strings.HasPrefix(customID, "release_add_modal|"):
		h.handleReleaseAddModal(s, i)
	case strings.HasPrefix(customID, "release_update_modal|"):
		h.handleReleaseUpdateModal(s, i)
	}
}
