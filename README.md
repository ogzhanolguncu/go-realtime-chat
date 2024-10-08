# Go Real-time Chat Client and Server

A feature-rich, real-time chat application implemented in Go, supporting both group and private messaging, with an intuitive terminal-based user interface.

## Features

- Real-time messaging in group chats and private conversations
- Channel system for topic-based discussions
- User authentication and registration
- Message history and active user list
- Blocking and muting functionalities
- Terminal-based UI with mouse support
- Sound notifications for new messages
- Customizable server configuration

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Architecture](#architecture)
- [Components](#components)
- [Configuration](#configuration)
- [Commands](#commands)
- [Contributing](#contributing)

## Installation

1. Ensure you have Go installed on your system (version 1.16 or later recommended).
2. Clone the repository:
   ```
   git clone https://github.com/yourusername/go-chat.git
   cd go-chat
   ```
3. Install dependencies:
   ```
   go mod tidy
   ```

## Usage

1. Start the server:
   ```
   go run server/main.go
   ```
2. In a separate terminal, start a client:
   ```
   go run client/main.go
   ```
3. Follow the on-screen instructions to log in or register.

## Architecture

The application follows a client-server architecture:

- **Server**: Manages connections, routes messages, and handles user authentication.
- **Client**: Provides a terminal-based UI for user interaction and communication.

## Components

### Server

- `main.go`: Entry point for the server application.
- `server.go`: Core server logic and connection handling.
- `message_router.go`: Routes messages between clients.
- `auth.go`: Handles user authentication and registration.
- `channels.go`: Manages chat channels and their operations.
- `chat_history.go`: Stores and retrieves message history.

### Client

- `main.go`: Entry point for the client application.
- `client.go`: Core client logic and server communication.
- `message_handler.go`: Processes incoming and outgoing messages.
- `ui_manager/`: Contains the terminal UI implementation.

### Shared

- `protocol/`: Defines the communication protocol between client and server.

## Configuration

The server can be configured using environment variables:

- `CHAT_PORT`: Port number for the server (default: 7007)

## Commands

Users can interact with the chat application using the following commands:

- `/whisper <username> <message>`: Send a private message
- `/reply <message>`: Reply to the last received private message
- `/mute <username>`: Mute messages from a user
- `/unmute <username>`: Unmute a previously muted user
- `/block <username>`: Block a user
- `/unblock <username>`: Unblock a previously blocked user
- `/ch create <name> <password> <max_users> <public|private>`: Create a new channel
- `/ch join <name> <password>`: Join an existing channel
- `/ch leave`: Leave the current channel
- `/ch users`: List users in the current channel
- `/ch list`: Show active channels
- `/ch kick <username>`: Kick a user from the channel (channel owner only)
- `/ch ban <username>`: Ban a user from the channel (channel owner only)
- `/clear`: Clear the chat screen
- `/quit`: Exit the application

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
