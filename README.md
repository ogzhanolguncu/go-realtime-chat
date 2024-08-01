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

### Chat Application Todo List:

    Core Functionality:
        ✔ Group message: Allow passing public messages
        ✔ Private message: Allow passing private messages with the format /w [name] [message]
        ✔ Make names unique: Ensure usernames are unique across all connected clients @done(24-07-10 09:51)
        ✔ Validate username: If username is less than 2 char or empty, retry 3 times
        ✔ Add timestamps to messages for better context
        ✔ Reply functionality: Implement reply functionality with /reply [message] @started(24-07-10 09:51) @done(24-07-10 10:11) @lasted(20m24s)
        ✔ Client re-establish connection with exponential retry @done(24-07-14 15:38)
    User Interface:
        ✔ Send messages from server with prefix (MSG, WSP)
        ✔ Display allowed commands when someone joins. This should be handled on client side. @started(24-07-15 00:04) @done(24-07-15 00:32) @lasted(28m41s)
        ☐ Client errors should have "Client:" prefix with a different color
        ✔ Implement Active User display: @done(24-07-15 15:44)
        ```
        Active Users (25): Alice, Bob, Charlie [+22 more] (Use /users for full list)
        ----------------------------------
        [Chat messages appear here]
        ----------------------------------
        Your message:
        ```

    Chat History:
        ✔ Add timestamp to all message types @done(24-07-15 23:43)
        ✔ Add new data structure to save history to memory
            - Implement a simple slice for general messages
            - Add a map for quick access to private messages
        ✔ Filter messages before dispatching them to clients
            - Delete private messages if user is not recipient or sender
            - Ensure private messages are not exposed to unintended recipients
        ✔ Store chat history for late-joining clients (limit to ~200 messages)
        ✔ Implement /history command
            - Return recent messages (up to 200)
            - Include only relevant private messages for the requesting user
        ✔ Client should format timestamps carefully because some messages are from past
            - Implement a function to format timestamps relative to current time
        ✔ Server should persist history to file and restore on restart
            - Implement SaveToFile and LoadFromFile functions
            - Only persist general and private messages, not system or error messages
            - Saved file should be named as chat_history_TIMESTAMP.txt. So we can easily check if its old enough to be deleted.
            - Should keep saving and checking on separate goroutine
        ☐ Delete persisted messages if they are older than a day
            - Add a cleanup function to remove old messages
    Protocol and Security:
        ✔ Add Protocol
        ✔ Implement protocol: Client read
        ✔ Implement protocol: Server sent
        ✔ Implement protocol: Client sent @done(24-07-07 12:38)
        ✔ Implement protocol: Server read @done(24-07-07 12:38)
        ☐ Message encryption: Implement end-to-end encryption using golang.org/x/crypto
    Advanced Features:
        Auth:
            - Persist everything on disk or redis.
            Presence System:
                ☐ Show user statuses (online, away, busy, etc.)
                ☐ Allow users to set custom status messages
        Chatrooms:
            ☐ Allow users to create and join different chatrooms
            ☐ Allow clients to switch between private chats and group chat
            ☐ Implement commands: /create, /join, /leave
            Typing Indicators (for private chats):
                ☐ Implement client-side typing detection
                    - Send "typing" signal when user starts typing in a private chat
                    - Send "stopped typing" signal after user pauses (e.g., 2 seconds of inactivity)
                ☐ Implement server-side handling of typing signals
                    - Forward typing status only to the relevant chat partner
                ☐ Add UI element for displaying typing status in private chats
                    - Show "[User] is typing..." below the last message in the chat
                    - Ensure it updates smoothly without disrupting the chat view
                ☐ Implement a timeout mechanism for typing indicators
                    - Automatically clear the typing indicator if no updates are received for a certain period (e.g., 5 seconds)
        Command System:
            ☐ Create a robust command system (e.g., /help, /mute, /unmute, /block)
                Help:
                    ✔ Client only. Show relevant commands @done(24-07-16 00:42)
                Unlock/Block:
                    ☐ When a client blocks someone, the client's messages won't be sent to or received by the blocked user.
                    ☐ Server shouldn't broadcast to blocked user. Similar whisper, but everyone except that user.
                    ☐ Blocked user's active users list should exclude blocker. For instance, if Oz blocks John, John shouldn't able to see Oz in active list.
                Mute:
                    ✔ Don't show muted users texts
        Notification System:
        ☐   Notify users of mentions or important messages when not actively viewing chat
        Active User Management:
            ✔ Server: Maintain list/map of all active users @done(24-07-16 00:53)
            ✔ Server: Send updates with total count and subset of names @done(24-07-16 00:53)
            ✔ Client: Store full list of users @done(24-07-16 00:53)
            ✔ Client: Implement /users command to show full list @done(24-07-16 00:53)
        Performance and Security:
            ☐ Implement rate limiting to prevent spam and abuse
    Bug Fixes:
        ✔ Fix: Server locks when two users join simultaneously
        ✔ Check whisper logic: Ensure robustness and correctness
        ✔ When client reconnects after successful connection it required one more extra input before sending username. @done(24-07-14 16:04)
            Make sure all goroutines are closed before retrying using waitGroups and additional channel called done @Solution
