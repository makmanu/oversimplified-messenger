# Oversimplified Messenger

A simple Go web messenger with a clean UI for sending and receiving messages. Messages are stored in a SQLite database.

## Features

- **Web Interface**: Simple HTML/CSS/JavaScript UI on port 50505
- **Send Messages**: Form to send messages with From, To, and Message fields
- **View Messages**: Search and view all messages for a specific recipient
- **Persistent Storage**: Messages stored in SQLite database
- **Docker Support**: Easy containerization with persistent volumes

## Running Locally

1. Install dependencies:
```bash
go mod download
go get github.com/mattn/go-sqlite3
```

2. Run the server:
```bash
go run main.go
```

The server will start on `http://localhost:50505`

## Running with Docker

Build and run with Docker Compose:
```bash
docker-compose up --build
```

The application will be available at `http://localhost:50505`

The database will be persisted in a Docker volume (`messenger_data`), so your messages will survive container restarts.

### Docker Compose Options

Stop the container:
```bash
docker-compose down
```

Restart the container:
```bash
docker-compose up
```

Remove the database volume:
```bash
docker-compose down -v
```

## API Endpoints

### Send Message
```bash
curl -X POST http://localhost:50505/messages \
  -H "Content-Type: application/json" \
  -d '{"from": "Alice", "to": "Bob", "message": "Hello Bob!"}'
```

Response (201 Created):
```json
{
  "id": 1,
  "from": "Alice",
  "to": "Bob",
  "message": "Hello Bob!",
  "status": "saved"
}
```

### Get Messages for a Recipient
```bash
curl http://localhost:50505/api/messages?to=Bob
```

Response:
```json
[
  {
    "id": 1,
    "from": "Alice",
    "to": "Bob",
    "message": "Hello Bob!",
    "created_at": "2024-04-03T10:30:45Z"
  }
]
```

## Database Schema

Messages are stored in SQLite with the following schema:
- `id`: Auto-incrementing primary key
- `from_user`: Message sender name
- `to_user`: Message recipient name
- `message`: Message content
- `created_at`: Timestamp of when the message was saved
