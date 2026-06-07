package model

// CreateChatRequest is the payload for creating a chat.
// For "direct" chats, exactly one other member ID is expected in MemberIDs.
// For "group" chats, a Name is required and any number of members may be added.
type CreateChatRequest struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	MemberIDs []string `json:"member_ids"`
}

// UpdateChatRequest is the payload for renaming a chat.
type UpdateChatRequest struct {
	Name *string `json:"name"`
}

// AddMemberRequest is the payload for adding a member to a chat.
type AddMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// UpdateMemberRoleRequest is the payload for changing a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}
