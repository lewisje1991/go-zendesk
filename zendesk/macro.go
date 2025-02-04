package zendesk

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Macro is information about zendesk macro
type Macro struct {
	Actions     []MacroAction `json:"actions"`
	Active      bool          `json:"active"`
	CreatedAt   time.Time     `json:"created_at,omitempty"`
	Description interface{}   `json:"description"`
	ID          int64         `json:"id,omitempty"`
	Position    int           `json:"position,omitempty"`
	Restriction interface{}   `json:"restriction"`
	Title       string        `json:"title"`
	UpdatedAt   time.Time     `json:"updated_at,omitempty"`
	URL         string        `json:"url,omitempty"`
}

// MacroAction is definition of what the macro does to the ticket
//
// ref: https://develop.zendesk.com/hc/en-us/articles/360056760874-Support-API-Actions-reference
type MacroAction struct {
	Field string   `json:"field"`
	Value []string `json:"value"`
}

// MacroListOptions is parameters used of GetMacros
type MacroListOptions struct {
	Access       string `json:"access"`
	Active       string `json:"active"`
	Category     int    `json:"category"`
	GroupID      int    `json:"group_id"`
	Include      string `json:"include"`
	OnlyViewable bool   `json:"only_viewable"`

	PageOptions

	// SortBy can take "created_at", "updated_at", "usage_1h", "usage_24h",
	// "usage_7d", "usage_30d", "alphabetical"
	SortBy string `url:"sort_by,omitempty"`

	// SortOrder can take "asc" or "desc"
	SortOrder string `url:"sort_order,omitempty"`
}

// MacroAPI an interface containing all macro related methods
type MacroAPI interface {
	GetMacros(ctx context.Context, opts *MacroListOptions) ([]Macro, Page, error)
	GetMacro(ctx context.Context, macroID int64) (Macro, error)
	CreateMacro(ctx context.Context, macro Macro) (Macro, error)
	UpdateMacro(ctx context.Context, macroID int64, macro Macro) (Macro, error)
	DeleteMacro(ctx context.Context, macroID int64) error
	ShowChangesToTicket(ctx context.Context, macroID int64) (Ticket, error)
	ShowTicketAfterChanges(ctx context.Context, ticketID, macroID int64) (Ticket, error)
}

// GetMacros get macro list
//
// ref: https://developer.zendesk.com/rest_api/docs/support/macros#list-macros
func (z *Client) GetMacros(ctx context.Context, opts *MacroListOptions) ([]Macro, Page, error) {
	var data struct {
		Macros []Macro `json:"macros"`
		Page
	}

	tmp := opts
	if tmp == nil {
		tmp = &MacroListOptions{}
	}

	u, err := addOptions("/macros.json", tmp)
	if err != nil {
		return nil, Page{}, err
	}

	body, err := z.get(ctx, u)
	if err != nil {
		return nil, Page{}, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, Page{}, err
	}
	return data.Macros, data.Page, nil
}

// GetMacro gets a specified macro
//
// ref: https://developer.zendesk.com/rest_api/docs/support/macros#show-macro
func (z *Client) GetMacro(ctx context.Context, macroID int64) (Macro, error) {
	var result struct {
		Macro Macro `json:"macro"`
	}

	body, err := z.get(ctx, fmt.Sprintf("/macros/%d.json", macroID))
	if err != nil {
		return Macro{}, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return Macro{}, err
	}

	return result.Macro, err
}

// CreateMacro create a new macro
//
// ref: https://developer.zendesk.com/rest_api/docs/support/macros#create-macro
func (z *Client) CreateMacro(ctx context.Context, macro Macro) (Macro, error) {
	var data, result struct {
		Macro Macro `json:"macro"`
	}
	data.Macro = macro

	body, err := z.post(ctx, "/macros.json", data)
	if err != nil {
		return Macro{}, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return Macro{}, err
	}
	return result.Macro, nil
}

// UpdateMacro update an existing macro
// ref: https://developer.zendesk.com/rest_api/docs/support/macros#update-macro
func (z *Client) UpdateMacro(ctx context.Context, macroID int64, macro Macro) (Macro, error) {
	var data, result struct {
		Macro Macro `json:"macro"`
	}
	data.Macro = macro

	path := fmt.Sprintf("/macros/%d.json", macroID)
	body, err := z.put(ctx, path, data)
	if err != nil {
		return Macro{}, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return Macro{}, err
	}

	return result.Macro, nil
}

// DeleteMacro deletes the specified macro
// ref: https://developer.zendesk.com/rest_api/docs/support/macros#delete-macro
func (z *Client) DeleteMacro(ctx context.Context, macroID int64) error {
	err := z.delete(ctx, fmt.Sprintf("/macros/%d.json", macroID))

	if err != nil {
		return err
	}

	return nil
}

// Returns the changes the macro would make to a ticket.
// It doesn't actually change a ticket. You can use the response data in a subsequent API call to the Tickets endpoint to update the ticket.
// ref: https://developer.zendesk.com/api-reference/ticketing/business-rules/macros/#show-changes-to-ticket
func (z *Client) ShowChangesToTicket(ctx context.Context, macroID int64) (Ticket, error) {
	body, err := z.get(ctx, fmt.Sprintf("/macros/%d/apply.json", macroID))
	if err != nil {
		return Ticket{}, err
	}

	unmarshal := func(data []byte) (Ticket, error) {
		type results struct {
			Result struct {
				Ticket struct {
					TicketFormID     string `json:"ticket_form_id"`
					SideConversation struct {
						Subject     string `json:"subject"`
						Message     string `json:"message"`
						Recipients  string `json:"recipients"`
						ContextType string `json:"context_type"`
					} `json:"side_conversation"`
					Subject string   `json:"subject"`
					Tags    []string `json:"tags"`
					Comment struct {
						Body   string `json:"body"`
						Public string `json:"public"`
					} `json:"comment"`
					CollaboratorIDs []int64       `json:"collaborator_ids"`
					FollowerIDs     []int64       `json:"follower_ids"`
					Status          string        `json:"status"`
					CustomFields    []CustomField `json:"custom_fields,omitempty"`
				} `json:"ticket"`
			} `json:"result"`
		}

		var r results
		err := json.Unmarshal(data, &r)
		if err != nil {
			return Ticket{}, err
		}

		commentIsPublic, err := strconv.ParseBool(r.Result.Ticket.Comment.Public)
		if err != nil {
			return Ticket{}, err
		}

		ticketFormId, err := strconv.ParseInt(r.Result.Ticket.TicketFormID, 10, 64)
		if err != nil {
			return Ticket{}, err
		}

		return Ticket{
			TicketFormID:     ticketFormId,
			SideConversation: r.Result.Ticket.SideConversation,
			Subject:          r.Result.Ticket.Subject,
			Tags:             r.Result.Ticket.Tags,
			Comment: &TicketComment{
				Body:   r.Result.Ticket.Comment.Body,
				Public: &commentIsPublic,
			},
			CollaboratorIDs: r.Result.Ticket.CollaboratorIDs,
			FollowerIDs:     r.Result.Ticket.FollowerIDs,
			Status:          r.Result.Ticket.Status,
		}, nil
	}

	//Zendesk api returns ticket.comment.public as string, not bool so needs custom unmarshalling
	return unmarshal(body)
}

// ShowTicketAfterChanges Returns the full ticket object as it would be after applying the macro to the ticket.
// It doesn't actually change the ticket.
// ref: https://developer.zendesk.com/api-reference/ticketing/business-rules/macros/#show-ticket-after-changes
func (z *Client) ShowTicketAfterChanges(ctx context.Context, ticketID, macroID int64) (Ticket, error) {
	body, err := z.get(ctx, fmt.Sprintf("/tickets/%d/macros/%d/apply", ticketID, macroID))
	if err != nil {
		return Ticket{}, err
	}

	unmarshal := func(data []byte) (Ticket, error) {
		type results struct {
			Result struct {
				Ticket struct {
					TicketFormID int64    `json:"ticket_form_id"`
					Subject      string   `json:"subject"`
					Tags         []string `json:"tags"`
					Comment      struct {
						Body   string `json:"body"`
						Public string `json:"public"`
					} `json:"comment"`
					CollaboratorIDs []int64       `json:"collaborator_ids"`
					FollowerIDs     []int64       `json:"follower_ids"`
					Status          string        `json:"status"`
					CustomFields    []CustomField `json:"custom_fields,omitempty"`
				} `json:"ticket"`
			} `json:"result"`
		}

		var r results
		err := json.Unmarshal(data, &r)
		if err != nil {
			return Ticket{}, err
		}

		commentIsPublic, err := strconv.ParseBool(r.Result.Ticket.Comment.Public)
		if err != nil {
			return Ticket{}, err
		}

		return Ticket{
			TicketFormID: r.Result.Ticket.TicketFormID,
			Subject:      r.Result.Ticket.Subject,
			Tags:         r.Result.Ticket.Tags,
			Comment: &TicketComment{
				Body:   r.Result.Ticket.Comment.Body,
				Public: &commentIsPublic,
			},
			CollaboratorIDs: r.Result.Ticket.CollaboratorIDs,
			FollowerIDs:     r.Result.Ticket.FollowerIDs,
			Status:          r.Result.Ticket.Status,
			CustomFields:    r.Result.Ticket.CustomFields,
		}, nil
	}

	//Zendesk api returns ticket.comment.public as string, not bool so needs custom unmarshalling
	return unmarshal(body)
}
