**Real-Time Forum**
A simple real-time forum built in Go with WebSocket-powered live updates and a minimal frontend.

Quick Start
Real-Time Forum
A simple real-time forum written in Go that demonstrates live updates, real-time presence, and a small single-page frontend. It is suitable as a learning project or a starting point for lightweight discussion apps.

What this project includes
Real-time messaging and presence using WebSockets (websocket/).
Persistent data access and queries in the database/ package (posts, comments, reactions, private messages, categories, users).
Session handling in the session/ package for basic user sessions.
Minimal frontend in static/ (HTML, CSS, and JavaScript) that interacts with the WebSocket backend for live updates.
Key Features
Create and list posts and comments
Real-time updates for new posts, comments, and reactions
Private messages between users
User presence (online/offline)
Reactions and counts on posts/comments
Category support for organizing posts
Quick Start
Prerequisites:

Go 1.18 or newer
A SQL database (Postgres, MySQL, or SQLite). Configure the connection for the database package before running.
Example environment variables (adjust to your setup):

PORT — server port (default 8080) $env:DB_DRIVER="postgres" Build a binary:
- `database/` - SQL schema helpers, queries, and DB initialization



Minimal Go forum with WebSocket-powered live updates.

## Quick Start

Prerequisites: Go 1.18+ and a configured SQL database.

Run locally:

```powershell
go run main.go
Open http://localhost:8080 in your browser.

Common env vars:

DB_DRIVER (e.g. postgres)

DB_DSN (database DSN). Example Postgres DSN:

postgres://user:pass@localhost:5432/forumdb?sslmode=disable

PORT (default 8080)

Layout
main.go
database/
websocket/
session/
static/ (HTML/CSS/JS)
Contributors
Hawra Fadhel
Khaireya Alhayki
