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

- [x] Private message: Allow passing private messages with the format /w [name] [message].
- [x] Reply functionality: Implement reply functionality with /r [message].
- [x] Make names unique: Ensure usernames are unique across all connected clients.
- [x] Group message: Allow passing public messages.
- Chat history
  - [] Implement chat history for clients who join late. After ~100 server should delete some of the messages.
  - [] Server should download history and restore when restarted
- [] Active people list: Display a list of active users.
- [] Client re-establish connection with exponential retry: Implement exponential retry logic for re-establishing connections.
- [] Send active user list: When a client establishes a connection, send them the list of active users.
- [] Check whisper logic: Ensure the logic for handling whispers is robust and correct.
- [x] Consider adding timestamps to messages for better context.
- [] When two users join at the same time server locks, which shouldn't
- [x] Send messages from server with prefix so client can decide how to show them like SYSTEM_MESSAGE, CHAT_MESSAGE(whisper and group)
  - [x] Add Protocol
  - [] Implement protocol
    - [x] Client read
    - [x] Server sent
    - [] Client sent
- [x] Validate username. If username is less than 2 char or empty. Retry 3 times.
