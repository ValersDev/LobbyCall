# LobbyCall

Summon the squad. Vote the mode. Lock the plan.

A Discord lobby bot for scheduling game nights without the group-chat chaos — built in **Go** with a clean handler → service → storage layout.

One open vote per channel, live tallies, optional role pings, and a custom message so the call actually sounds like your crew.

---

## What it does

| Command / action | Result |
| --- | --- |
| `/planificar hora:22:30` | Opens a lobby with vote buttons |
| `mensaje:` *(optional)* | Adds a custom note to the post |
| Vote buttons | Flex / 5v5 / ARAM / Chill — one vote per user, switch anytime |
| **Cerrar votación** or `/cerrar` | Locks the lobby, drops buttons, announces the winner |

Also:
- Validates `HH:MM`
- Shows who’s leading while voting
- Pings a configured role on open + close
- Enforces **one open plan per channel**

---

## Stack

- **Go** + [discordgo](https://github.com/bwmarrin/discordgo)
- In-memory store with mutex-safe vote updates
- Docker / Compose ready for always-on deploy

```
cmd/bot                  entrypoint
internal/
  handlers/              Discord interactions
  services/              lobby rules + validation
  storage/               in-memory persistence
  models/                Plan / Player
  enums/                 commands, options, vote IDs
  config/                env loading
  errs/                  shared sentinel errors
```

---

## Quick start

**Requirements:** Go 1.22+ (developed on 1.27), a Discord bot token, guild ID.

```bash
cp .env.example .env
go build -o bot ./cmd/bot
./bot
```

On Windows (if `go run` is blocked by Application Control):

```powershell
go build -o bot.exe ./cmd/bot
.\bot.exe
```

### Environment

| Variable | Required | Description |
| --- | --- | --- |
| `DISCORD_TOKEN` | yes | Bot token |
| `GUILD_ID` | yes | Server ID (guild-scoped commands) |
| `LOBBY_ROLE_ID` | no | Role to ping on plan open/close |

```env
DISCORD_TOKEN=...
GUILD_ID=...
LOBBY_ROLE_ID=...
```

Invite the bot with `bot` + `applications.commands`. Enable Developer Mode in Discord to copy IDs.

---

## Docker

```bash
docker compose up -d --build
docker compose logs -f
```

---

## Design notes

- Votes update under a write lock — no lost tallies under concurrent clicks
- Slash commands stay thin; business rules live in `LobbyService`
- CustomIDs and command names are centralized so Discord strings don’t leak everywhere
- Memory store keeps the deploy simple; swap the storage layer later without rewriting handlers

---

## Roadmap

- [ ] Persist lobbies across restarts (SQLite)
- [ ] Deduped slash-command sync on startup
- [ ] Richer close flow (reminders, voice channel link)

---

## License

**Source-available** under [PolyForm Noncommercial 1.0.0](LICENSE).

You can read, study, fork, and run LobbyCall for personal / non-commercial use.  
You **may not** sell it, resell it, or offer it as a hosted product or paid service.

If you need commercial rights, ask for a separate license.
