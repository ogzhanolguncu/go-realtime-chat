package protocol

// Group/General Message (MSG): MSG|sender|message_length|message_content\r\n
// Whisper/DM Message (WSP): 	WSP|recipient|message_length|message_content\r\n
// System Notice (SYS): 		SYS|message_length|message_content|status \r\n status = "fail" | "success"
// Error Message (ERR): 		ERR|message_length|error_message\r\n
const Separator = "|"

const (
	MessageTypeMSG = "MSG"
	MessageTypeWSP = "WSP"
	MessageTypeSYS = "SYS"
	MessageTypeERR = "ERR"
)

type Payload struct {
	Content     string
	ContentType string
	Sender      string
	SysStatus   string
}
