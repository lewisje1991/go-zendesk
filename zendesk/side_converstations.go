package zendesk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type SideConversation struct {
	CreatedAt      time.Time      `json:"created_at,omitempty"`
	ID             string         `json:"id,omitempty"`
	MessageAddedAt time.Time      `json:"message_added_at,omitempty"`
	Participants   []Participants `json:"participants,omitempty"`
	PreviewText    string         `json:"preview_text,omitempty"`
	State          string         `json:"state,omitempty"`
	StateUpdatedAt time.Time      `json:"state_updated_at,omitempty"`
	Subject        string         `json:"subject,omitempty"`
	TicketID       int64          `json:"ticket_id,omitempty"`
	UpdatedAt      time.Time      `json:"updated_at,omitempty"`
	URL            string         `json:"url,omitempty"`
}

type Message struct {
	Subject     string            `json:"subject,omitempty"`
	PreviewText string            `json:"preview_text,omitempty"`
	Body        string            `json:"body,omitempty"`
	HTMLBody    string            `json:"html_body,omitempty"`
	From        map[string]string `json:"from,omitempty"`
	To          []MessageTo       `json:"to,omitempty"`
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
}

type Participants struct {
	UserID int64  `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
}

type ExternalIDs struct {
	MySystemID string `json:"my_system_id,omitempty"`
}

type MessageTo struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// CreateSideConversation create a new side conversation
//
// ref: https://developer.zendesk.com/api-reference/ticketing/side_conversation/side_conversation/#create-side-conversation
func (z *Client) CreateSideConversation(ctx context.Context, ticketID int64, m Message) (SideConversation, error) {
	var request struct {
		Message Message `json:"message"`
	}
	request.Message = m

	body, err := z.post(ctx, fmt.Sprintf("/tickets/%d/side_conversations", ticketID), request)
	if err != nil {
		return SideConversation{}, err
	}

	var result struct {
		SideConversation SideConversation `json:"side_conversation"`
	}

	fmt.Println(string(body))

	err = json.Unmarshal(body, &result)
	if err != nil {
		return SideConversation{}, err
	}
	return result.SideConversation, nil
}
