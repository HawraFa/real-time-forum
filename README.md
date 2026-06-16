# Real-Time Forum

A simple real-time forum built in Go that demonstrates live updates, real-time presence, and a minimal single-page frontend. It is suitable as a learning project or a starting point for lightweight discussion applications.

## Features

- Create and list posts and comments  
- Real-time updates for posts, comments, and reactions (WebSockets)  
- Private messaging between users  
- User presence (online/offline status)  
- Reactions and counters on posts/comments  
- Category support for organizing posts  

## Project Structure

main.go  
database/        # DB schema, queries, initialization  
websocket/       # WebSocket connection and real-time logic  
session/         # User session management  
static/          # Frontend (HTML, CSS, JavaScript)  

## Requirements

- Go 1.18 or newer  
- SQL database (PostgreSQL, MySQL, or SQLite)  

## Environment Variables

PORT=8080  
DB_DRIVER=postgres  
DB_DSN=postgres://user:password@localhost:5432/forumdb?sslmode=disable  

## Quick Start

Clone the repository:

git clone https://github.com/HawraFa/forum.git  
cd forum  

Run the project:

go run main.go  

Open in browser:

http://localhost:8080  

## Build Binary (Optional)

go build -o forum  

## Core Modules

database/  
SQL schema, queries, and database initialization  

websocket/  
Real-time communication layer  

session/  
Authentication and session handling  

static/  
Frontend interface (HTML, CSS, JavaScript)  

## Contributors

Hawra Fadhel  
Khaireya Alhayki  

## Notes

This project demonstrates:
- WebSocket-based real-time systems  
- Backend architecture in Go  
- Simple single-page application integration  
