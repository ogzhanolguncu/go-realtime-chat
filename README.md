## Real-time Chat Server and Client

This project implements a simple real-time chat server and client using Go. The current implementation supports:

- Group messages
- Private messages (whispers)
- Usernames
- Color-coded messages for easy readability

### Features

- Server

  - Group Messaging: Broadcast messages to all connected clients.
  - Private Messaging (Whisper): Send direct messages to specific users.
  - Username Management: Ensures that each user sets a username upon connection.
  - Color-coded Messages: Assigns a unique color to each user's messages for easy differentiation.
  - Concurrent Connections: Handles multiple client connections concurrently.

- Client

  - Connect to Server: Establish a connection to the chat server.
  - Set Username: Prompt to set a username upon connection.
  - Send Messages: Send group messages or private messages to other users.
  - Receive Messages: Display incoming messages from the server.

### TODOs

{% include_relative ./tasks.todo %}
