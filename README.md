# 🚀 Release Bot

A Discord bot for managing weekly release threads — built in Go using [`discordgo`](https://github.com/bwmarrin/discordgo).

PMs initialize a release, developers add and update their branches, and the bot keeps a live summary message in the thread updated automatically.

---

## Features

- 📋 Create release threads with one command
- ➕ Add branch entries with status, PR link, and blockers
- ✏️ Update your own entries via guided dropdowns
- ❌ Remove a branch from the release (shown struck-through in the thread, excluded from summary counts)
- 📊 Instant ephemeral summary of release status
- 🔒 Thread-safe in-memory state

---

## Commands

| Command | Who | What it does |
|---|---|---|
| `/ping` | Anyone | Health check — responds `Pong 🚀` |
| `/release-init` | PM | Starts a new release (type → date → notes → thread) |
| `/release-add` | Developer | Adds a branch entry to the active release |
| `/release-update` | Developer | Updates a branch entry — or removes it from the release |
| `/release-summary` | Anyone | Shows an ephemeral count snapshot by status (removed entries excluded) |

---

## Setup Guide

### Prerequisites

- [Go](https://go.dev/dl/) installed (`go 1.21+` recommended)
- A Discord account
- Permission to add bots to the target server

---

### 1. Clone the repository

```bash
git clone https://github.com/abhinav2712/release-bot.git
cd release-bot
```

---

### 2. Create a Discord application

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **New Application** and give it a name
3. Navigate to the **Bot** section
4. Click **Add Bot** → confirm
5. Under **Token**, click **Reset Token** and copy it

> This is your `DISCORD_TOKEN`. Keep it secret.

---

### 3. Invite the bot to your server

1. In the Developer Portal, open **OAuth2 → URL Generator**
2. Select scopes:
   - ✅ `bot`
   - ✅ `applications.commands`
3. Select bot permissions:
   - ✅ View Channels
   - ✅ Send Messages
   - ✅ Read Message History
   - ✅ Manage Threads
4. Copy the generated URL, open it in your browser, and add the bot to your server

---

### 4. Get your Guild ID (Server ID)

1. Open Discord
2. Go to **User Settings → Advanced**
3. Enable **Developer Mode**
4. Right-click your server name in the sidebar
5. Click **Copy Server ID**

> This is your `GUILD_ID`.

---

### 5. Set up environment variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

```env
DISCORD_TOKEN=your_bot_token_here
GUILD_ID=your_server_id_here
```

| Variable | Description |
|---|---|
| `DISCORD_TOKEN` | Bot token from the Discord Developer Portal |
| `GUILD_ID` | ID of the server where slash commands are registered |

---

### 6. Install dependencies

```bash
go mod tidy
```

---

### 7. Run the bot

```bash
go run cmd/bot/main.go
```

If everything is configured correctly, you'll see:

```
Bot is ready!
Bot is running...
```

---

### 8. Test it out

In your Discord server, run:

```
/ping
```

If the bot replies with `Pong 🚀`, you're good to go.

---

## Release Flow

```
PM runs /release-init
  └─ selects release type (major / minor)
  └─ selects release date
  └─ enters release notes in modal
  └─ bot creates a thread + posts a live summary message

Developers run /release-add
  └─ select a status
  └─ fill in branch, title, PR link, blocker
  └─ summary message updates automatically

Developers run /release-update
  └─ select their branch
  └─ select new status
        ├─ normal status → fill in updated PR / blocker in modal → summary updates
        └─ "Remove from Release" → no modal, branch instantly struck through in thread

Anyone runs /release-summary
  └─ receives an ephemeral count breakdown (removed branches excluded)
```

---

## Project Structure

```
release-bot/
├── cmd/bot/main.go               # Entry point — env, session wiring, signal handling
└── internal/
    ├── models/release.go         # ReleaseItem and CurrentRelease types
    ├── state/store.go            # Thread-safe in-memory state store
    ├── status/status.go          # Status helpers, emoji map, date options
    ├── discord/respond.go        # Interaction response helpers
    ├── summary/builder.go        # Summary message builder
    └── handlers/
        ├── router.go             # Top-level interaction dispatcher
        ├── release_init.go       # /release-init flow
        ├── release_add.go        # /release-add flow
        ├── release_update.go     # /release-update flow
        └── release_summary.go    # /release-summary
```

---

## Status Reference

| Status | Emoji | Meaning |
|---|---|---|
| `in-progress` | 🛠️ | Work is ongoing |
| `given-for-review` | 👀 | Shared for review |
| `reviewed` | ✅ | Review is complete |
| `tested` | 🧪 | Testing is complete |
| `reviewed-and-tested` | 🚀 | Ready to ship |
| `removed` | ❌ | Removed from the release — struck-through in thread, excluded from counts |

> [!NOTE]
> The `removed` status is only available via `/release-update`. It cannot be set when first adding an entry with `/release-add`.

---

## Troubleshooting

**Slash commands are not showing up**
- Make sure the bot was invited with `applications.commands` scope
- Verify `GUILD_ID` matches the correct server
- Confirm the bot is running

**"Missing Access" error**
- The bot is not in the server, or was invited without the correct scopes
- Double-check the invite URL was generated with both `bot` and `applications.commands`

**`.env` values are not being read**
- The `.env` file must be in the project root, the same directory you run the command from

---

## Security

> [!CAUTION]
> Never commit your real `.env` file or bot token to version control.
> Only commit `.env.example` with empty values.
> If your token is ever exposed, reset it immediately from the [Discord Developer Portal](https://discord.com/developers/applications).

---

## License

MIT