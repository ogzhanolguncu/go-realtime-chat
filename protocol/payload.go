package protocol

// Group/General Message (MSG): 	MSG|timestamp|sender|message_length|message_content\r\n
// Whisper/DM Message (WSP): 		WSP|timestamp|sender|recipient|message_length|message_content\r\n
// System Notice (SYS): 			SYS|timestamp|message_length|message_content|status \r\n status = "fail" | "success"
// Error Message (ERR): 			ERR|timestamp|message_length|error_message\r\n
// Active Users(USR):				ACT_USRS|timestamp|active_user_length|active_user_array|status\r\n status = "res" | "req"
// Username Message(ACT_USRS): 		USR|timestamp|name_length|name_content|status\r\n status = "fail | "success"
// Chat History(HSTRY): 			HSTRY|timestamp|requester|messages_array|status\r\n status = "res" | "req"
const Separator = "|"

type MessageType string

const (
	MessageTypeMSG      MessageType = "MSG"
	MessageTypeWSP      MessageType = "WSP"
	MessageTypeSYS      MessageType = "SYS"
	MessageTypeERR      MessageType = "ERR"
	MessageTypeUSR      MessageType = "USR"
	MessageTypeACT_USRS MessageType = "ACT_USRS" //Active users
	MessageTypeHSTRY    MessageType = "HSTRY"    //Chat history
)

type Payload struct {
	Timestamp   int64
	Content     string
	MessageType MessageType
	Sender      string
	Recipient   string
	Status      string
	Username    string
	ActiveUsers []string

	EncodedChatHistory []string // Comma separated messages
	DecodedChatHistory []Payload
}
