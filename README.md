# ironsworn-backend

A Go backend for the [Ironsworn](https://www.ironswornrpg.com/) tabletop RPG (solo play mode). Provides a REST API for game state management, an MCP server for LLM integration, and a structured game log designed for RAG corpus generation.

Ironsworn is created by Shawn Tomkin and licensed under [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/).

## Features

- **33 Ironsworn moves** — all adventure, quest, combat, suffer, relationship, and fate moves
- **19 oracle tables** — action, theme, region, location, settlement, character, combat, and fate tables from the SRD
- **REST API** — 22 endpoints for full game management
- **MCP Server** — 15 tools and 5 resources for LLM-driven play (Claude Desktop, etc.)
- **RAG-ready game log** — before/after character snapshots, tags, JSONL export for embedding pipelines
- **Pure Go** — no CGO required (uses modernc.org/sqlite)

## Quick Start

```bash
# Run the REST API
go run main.go

# Run the MCP server (stdio, for Claude Desktop)
go run cmd/mcp/main.go

# Run tests
go test ./...
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `IRONSWORN_PORT` | `8080` | REST API port |
| `IRONSWORN_DB_PATH` | `ironsworn.db` | SQLite database path |

## Architecture

```
┌─────────────┐   ┌─────────────┐
│  REST API    │   │  MCP Server │
│  (Chi/HTTP)  │   │  (stdio)    │
└──────┬───────┘   └──────┬──────┘
       │                  │
       ▼                  ▼
┌─────────────────────────────────┐
│         Handler Layer           │
│  (orchestration: load→exec→save)│
└──────────────┬──────────────────┘
               │
       ┌───────┴────────┐
       ▼                ▼
┌────────────┐  ┌─────────────┐
│   Engine   │  │   Oracle    │
│ (pure game │  │  (tables +  │
│   logic)   │  │   rolls)    │
└────────────┘  └─────────────┘
               │
               ▼
┌─────────────────────────────────┐
│     Store (SQLite + RAG Log)    │
└─────────────────────────────────┘
```

## REST API

### Games
```
POST   /api/games                              Create game
GET    /api/games                              List games
GET    /api/games/{gameID}                     Get game
DELETE /api/games/{gameID}                     Delete game
```

### Character
```
POST   /api/games/{gameID}/character           Create character
GET    /api/games/{gameID}/character           Get character state
PATCH  /api/games/{gameID}/character           Manual adjustments
```

### Moves
```
POST   /api/games/{gameID}/moves              Execute a move
POST   /api/games/{gameID}/burn-momentum      Burn momentum
GET    /api/moves                              List all moves
GET    /api/moves/{moveID}                     Move details/rules
```

### Progress Tracks
```
POST   /api/games/{gameID}/progress           Create track
GET    /api/games/{gameID}/progress           List tracks
PATCH  /api/games/{gameID}/progress/{id}      Mark progress
DELETE /api/games/{gameID}/progress/{id}      Remove track
```

### Oracle
```
POST   /api/games/{gameID}/oracle/ask         Yes/no question
POST   /api/games/{gameID}/oracle/roll        Roll on table
GET    /api/oracle/tables                      List tables
```

### Assets
```
POST   /api/games/{gameID}/character/assets   Add asset
GET    /api/games/{gameID}/character/assets   List assets
PATCH  /api/games/{gameID}/character/assets/{id}  Upgrade ability
```

### Game Log (RAG)
```
GET    /api/games/{gameID}/log                Paginated log (filters: type, tags, outcome)
GET    /api/games/{gameID}/log/export         JSONL bulk export
POST   /api/games/{gameID}/log                Add narrative entry
```

## MCP Server

For use with Claude Desktop or any MCP-compatible client. Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "ironsworn": {
      "command": "go",
      "args": ["run", "/path/to/Ironsworn/cmd/mcp/main.go"]
    }
  }
}
```

### Tools

| Tool | Description |
|------|-------------|
| `create_game` | Create a new campaign |
| `create_character` | Create a character with stat distribution |
| `get_character` | Get current character state |
| `make_move` | Execute any Ironsworn move |
| `burn_momentum` | Burn momentum to improve last roll |
| `list_vows` | List active vows with progress |
| `create_vow` | Swear an iron vow |
| `mark_progress` | Mark progress on a track |
| `ask_oracle` | Ask a yes/no question with likelihood |
| `roll_oracle_table` | Roll on a named oracle table |
| `get_game_log` | Retrieve recent game log entries |
| `add_narrative` | Add narrative text to the game log |
| `list_moves` | List all available moves |
| `get_move_details` | Get rules for a specific move |

### Resources

| Resource | Description |
|----------|-------------|
| `ironsworn://character/{gameID}` | Current character state |
| `ironsworn://vows/{gameID}` | Active vows and progress |
| `ironsworn://log/{gameID}` | Recent game log |
| `ironsworn://rules/moves/{moveID}` | Move rules reference |
| `ironsworn://oracle/tables` | Available oracle tables |

## Example: Complete Vow Loop

```bash
BASE=http://localhost:8080/api

# Create a game
GAME=$(curl -s -X POST $BASE/games -d '{"name":"The Ironlands"}')
GAME_ID=$(echo $GAME | jq -r .id)

# Create a character (stats distributed as 3,2,2,1,1)
curl -s -X POST $BASE/games/$GAME_ID/character \
  -d '{"name":"Kara","stats":{"edge":2,"heart":3,"iron":1,"shadow":2,"wits":1}}'

# Create a vow
VOW=$(curl -s -X POST $BASE/games/$GAME_ID/progress \
  -d '{"name":"Protect Thornwall","track_type":"vow","rank":"dangerous"}')
VOW_ID=$(echo $VOW | jq -r .id)

# Swear the vow
curl -s -X POST $BASE/games/$GAME_ID/moves \
  -d '{"move_id":"swear_an_iron_vow","narrative":"I swear to protect the people of Thornwall."}'

# Reach milestones
curl -s -X POST $BASE/games/$GAME_ID/moves \
  -d "{\"move_id\":\"reach_a_milestone\",\"track_id\":\"$VOW_ID\"}"

# Fulfill the vow
curl -s -X POST $BASE/games/$GAME_ID/moves \
  -d "{\"move_id\":\"fulfill_your_vow\",\"track_id\":\"$VOW_ID\"}"

# Export game log as JSONL (for RAG pipelines)
curl -s $BASE/games/$GAME_ID/log/export > game_log.jsonl
```

## Game Mechanics

### Dice

- **Action Roll**: 1d6 + stat + adds vs 2d10 — Strong Hit / Weak Hit / Miss
- **Progress Roll**: progress score (0-10) vs 2d10 — no action die, can't burn momentum
- **Oracle Roll**: d100 for table lookups and yes/no questions
- **Matches**: when both challenge dice show the same value, something extraordinary happens

### Character Stats

Five stats distributed as 3, 2, 2, 1, 1: Edge, Heart, Iron, Shadow, Wits

### Progress Tracks

Stored as ticks (0-40). Progress score = ticks / 4 (0-10).

| Rank | Ticks per Mark |
|------|---------------|
| Troublesome | 12 |
| Dangerous | 8 |
| Formidable | 4 |
| Extreme | 2 |
| Epic | 1 |

## Dependencies

| Package | Purpose |
|---------|---------|
| [go-chi/chi](https://github.com/go-chi/chi) | HTTP routing |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | Pure Go SQLite driver |
| [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) | MCP server SDK |

## Project Structure

```
├── main.go                          # REST API entry point
├── cmd/mcp/main.go                  # MCP server entry point
├── internal/
│   ├── config/                      # Server configuration
│   ├── model/                       # Data types
│   ├── engine/                      # Pure game logic (no I/O)
│   │   ├── dice.go                  # Roller interface, action/progress rolls
│   │   ├── moves.go                 # Move registry and dispatcher
│   │   ├── adventure.go             # Face Danger, Gather Information, etc.
│   │   ├── quest.go                 # Swear an Iron Vow, Fulfill, etc.
│   │   ├── combat.go                # Strike, Clash, End the Fight, etc.
│   │   ├── suffer.go                # Endure Harm/Stress, Face Death, etc.
│   │   ├── relationship.go          # Compel, Sojourn, Forge a Bond, etc.
│   │   ├── fate.go                  # Pay the Price, Ask the Oracle
│   │   ├── momentum.go              # Burn, reset, negative cancel
│   │   └── progress.go              # Mark progress, track creation
│   ├── oracle/                      # Oracle engine + 19 embedded tables
│   ├── store/                       # SQLite persistence layer
│   ├── handler/                     # HTTP handlers
│   ├── gamelog/                      # RAG-ready structured logger
│   ├── mcpserver/                   # MCP tool/resource definitions
│   └── router/                      # Chi router setup
├── data/                            # Static data files
└── test/testutil/                   # Test helpers
```

## License

The code in this repository is available under the MIT License. Ironsworn game content is used under [Creative Commons Attribution 4.0 International](https://creativecommons.org/licenses/by/4.0/) — created by Shawn Tomkin.
