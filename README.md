# CodeAct Agent - KDrama Tracker

An AI agent that receives natural language commands and generates executable Go code to manage a KDrama database. Built with the **code-as-action** paradigm: instead of mapping commands to fixed functions, the LLM writes and executes real Go programs on every request.

## What it does

A web-based KDrama tracker where an AI agent performs CRUD operations, exports, stats, and more — all by writing code at runtime. Supports commands in English and Spanish.

**Capabilities:**

- Add dramas (with title, genre, rating, status)
- List all dramas or filter by rating/genre/status
- Update drama status (Watching, Completed, Dropped, Plan to Watch, Rewatching)
- Delete dramas (single or bulk by condition)
- Export the full database to JSON (downloadable file)
- Show stats (count, average rating, by genre/status)
- Health check (verify DB connection)
- Generate Go boilerplate code (e.g. HTTP handlers)
- Multi-provider support: switch between Ollama (local) and Groq (cloud) via env var
- Self-correcting: if generated code fails, the agent retries with the error context (up to 3 attempts)

## How it uses Code-as-Action

1. User sends a natural language command (e.g. "add Vincenzo with rating 9")
2. The LLM generates a **complete Go program** that uses GORM to query/modify a SQLite database
3. The program is saved to a temp file and executed with `go run`
4. If execution fails, the error is fed back to the LLM to self-correct (up to 3 retries)

The agent doesn't call predefined tools — it writes arbitrary code each time, giving it full access to loops, conditionals, multi-step logic, and error handling in a single action.

## Build & Run

**Prerequisites:** Go 1.21+, and either [Ollama](https://ollama.com) (local) or a [Groq](https://console.groq.com) API key (cloud, free tier).

```bash
git clone https://github.com/smgarciagr/codeact-agent-drama.git
cd codeact-agent-drama
cp .env.example .env   # edit with your provider choice
go run main.go         # opens at http://localhost:3000
```

`.env` configuration:

```env
# Option A: local (requires ollama pull llama3.2)
LLM_PROVIDER=ollama

# Option B: cloud
LLM_PROVIDER=groq
GROQ_API_KEY=your_key_here
```
