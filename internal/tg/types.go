package tg

// https://core.telegram.org/bots/api#update
type Update struct {
	UpdateID        int64            `json:"update_id"`
	Message         *Message         `json:"message,omitempty"`
	ChatJoinRequest *ChatJoinRequest `json:"chat_join_request,omitempty"`
}

// https://core.telegram.org/bots/api#message
type Message struct {
	MessageID int64  `json:"message_id"`
	From      User   `json:"from,omitempty"`
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text,omitempty"`
}

// https://core.telegram.org/bots/api#chatjoinrequest
type ChatJoinRequest struct {
	Chat       Chat  `json:"chat"`
	From       User  `json:"from"`
	UserChatID int64 `json:"user_chat_id"`
}

// https://core.telegram.org/bots/api#chat
type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
}

// https://core.telegram.org/bots/api#user
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// https://core.telegram.org/bots/api#chatmember
type ChatMember struct {
	Status string `json:"status"`
	User   *User  `json:"user"`
}
