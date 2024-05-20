package mailbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type MB struct {
	token string
}

type Mailbox struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type MailboxResponse struct {
	Embedded struct {
		Mailboxes []Mailbox `json:"mailboxes"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

type Folder struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type FoldersResponse struct {
	Embedded struct {
		Folders []Folder `json:"folders"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

type Conversation struct {
	ID        int    `json:"id"`
	Threads   int    `json:"threads"`
	Subject   string `json:"subject"`
	CreatedBy struct {
		Email string `json:"email"`
	} `json:"createdBy"`
	MailboxID       int `json:"mailboxId"`
	FolderID        int `json:"folderId"`
	PrimaryCustomer struct {
		Email string `json:"email"`
	} `json:"primaryCustomer"`
	Embedded struct {
		Threads []struct {
			Type   string `json:"type"`
			Action struct {
				Type string `json:"type"`
				Text string `json:"text,omitempty"`
			} `json:"action"`
			CreatedBy struct {
				Email string `json:"email"`
			} `json:"createdBy"`
			CreatedAt string `json:"createdAt"`
			Body      string `json:"body,omitempty"`
			Embedded  struct {
				Attachments []struct {
					ID    int    `json:"id"`
					Name  string `json:"name"`
					Mime  string `json:"mimeType"`
					Size  int    `json:"size"`
					URL   string `json:"url"`
					Links struct {
						Web struct {
							Href string `json:"href"`
						} `json:"web"`
					} `json:"_links"`
				} `json:"attachments,omitempty"`
			} `json:"_embedded"`
		} `json:"threads"`
	} `json:"_embedded"`
	CreatedAt string `json:"createdAt"`
}

type ConversationsResponse struct {
	Embedded struct {
		Conversations []Conversation `json:"conversations"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

func NewMailbox(token string) *MB {
	return &MB{token}
}

func (m *MB) MakeRequest(method string, endpoint string) *http.Response {
	client := &http.Client{}

	req, err := http.NewRequest(method, "https://api.helpscout.net/v2/"+endpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	req.Header.Add("Authorization", "Bearer "+m.token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}

	return resp
}

func (m *MB) GetMailboxes(page int) []Mailbox {
	resp := m.MakeRequest("GET", "mailboxes?page="+strconv.Itoa(page))
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var mailboxResponse MailboxResponse
	if err := json.Unmarshal(body, &mailboxResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil
	}

	mbs := mailboxResponse.Embedded.Mailboxes

	if page <= mailboxResponse.Page.TotalPages {
		results := m.GetMailboxes(page + 1)
		mbs = append(mbs, results...)
	}

	return mbs
}

func (m *MB) GetFolders(mailboxId int, page int) []Folder {
	resp := m.MakeRequest(
		"GET",
		"mailboxes/"+strconv.Itoa(mailboxId)+"/folders?page="+strconv.Itoa(page),
	)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var foldersResponse FoldersResponse
	if err := json.Unmarshal(body, &foldersResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil
	}

	folders := foldersResponse.Embedded.Folders

	if page < foldersResponse.Page.TotalPages {
		results := m.GetFolders(mailboxId, page+1)
		folders = append(folders, results...)
	}

	return folders
}

func (m *MB) GetConversations(page int) []Conversation {
	resp := m.MakeRequest("GET", "conversations?status=all&embed=threads&page="+strconv.Itoa(page))
	// resp := m.MakeRequest("GET", "conversations?embed=threads&page="+strconv.Itoa(page))
	fmt.Println("Fetching conversations?embed=threads&page=" + strconv.Itoa(page))
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var conversationsResponse ConversationsResponse
	if err := json.Unmarshal(body, &conversationsResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil
	}

	conversations := conversationsResponse.Embedded.Conversations

	if page < conversationsResponse.Page.TotalPages {
		results := m.GetConversations(page + 1)
		conversations = append(conversations, results...)
	}

	return conversations
}
