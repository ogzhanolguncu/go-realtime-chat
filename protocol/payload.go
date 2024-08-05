package protocol

// Group/General Message (MSG): 	MSG|timestamp|sender|message_content\r\n
// Whisper/DM Message (WSP): 		WSP|timestamp|sender|recipient|message_content\r\n
// System Notice (SYS): 			SYS|timestamp|message_content|status \r\n status = "fail" | "success"
// Error Message (ERR): 			ERR|timestamp|error_message\r\n
// Active Users(ACT_USRS):			ACT_USRS|timestampactive_user_array|status\r\n status = "res" | "req"
// Username Message(USR): 			USR|timestamp|username|password|status\r\n status = "fail | "success"
// Chat History(HSTRY): 			HSTRY|timestamp|requester|messages_array|status\r\n status = "res" | "req"
// Encryption(ENC): 				ENC|timestamp|requester_public_key|encrypted_group_chat_key
const Separator = "|"

type MessageType string

const (
	MessageTypeMSG      MessageType = "MSG"
	MessageTypeWSP      MessageType = "WSP"
	MessageTypeSYS      MessageType = "SYS"
	MessageTypeERR      MessageType = "ERR"
	MessageTypeUSR      MessageType = "USR"
	MessageTypeBLCK_USR MessageType = "BLCK_USR"
	MessageTypeACT_USRS MessageType = "ACT_USRS" //Active users
	MessageTypeHSTRY    MessageType = "HSTRY"    //Chat history
	MessageTypeENC      MessageType = "ENC"
)

type Payload struct {
	Timestamp   int64
	Content     string
	MessageType MessageType
	Sender      string
	Recipient   string
	Status      string

	Username string
	Password string

	ActiveUsers []string

	EncodedChatHistory []string // Comma separated messages
	DecodedChatHistory []Payload

	EncryptedKey string
}
