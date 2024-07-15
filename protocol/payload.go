package protocol

// Group/General Message (MSG): MSG|sender|message_length|message_content\r\n
// Whisper/DM Message (WSP): 	WSP|sender|recipient|message_length|message_content\r\n
// System Notice (SYS): 		SYS|message_length|message_content|status \r\n status = "fail" | "success"
// Error Message (ERR): 		ERR|message_length|error_message\r\n
// Username Message: 			USR|name_length|name_content|status\r\n status = "fail | "success"
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
	Content     string
	MessageType MessageType
	Sender      string
	Recipient   string
	Status      string
	Username    string
	ActiveUsers []string
}
