package protocol

// Group/General Message (MSG): MSG|timestamp|sender|message_length|message_content\r\n
// Whisper/DM Message (WSP): 	WSP|timestamp|sender|recipient|message_length|message_content\r\n
// System Notice (SYS): 		SYS|timestamp|message_length|message_content|status \r\n status = "fail" | "success"
// Error Message (ERR): 		ERR|timestamp|message_length|error_message\r\n
// Active Users:				ACT_USRS|timestamp|active_user_length|active_user_array|status\r\n status = "req" | "res"
// Username Message: 			USR|timestampname_length|name_content|status\r\n status = "fail | "success"
const Separator = "|"

type MessageType string

const (
	MessageTypeMSG      MessageType = "MSG"
	MessageTypeWSP      MessageType = "WSP"
	MessageTypeSYS      MessageType = "SYS"
	MessageTypeERR      MessageType = "ERR"
	MessageTypeUSR      MessageType = "USR"
	MessageTypeACT_USRS MessageType = "ACT_USRS" //Active users
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
}
