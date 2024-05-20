package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/roks0n/helpscout-archiver/internal/mailbox"
)

type MemoryStorage struct {
	Mailboxes map[int]string
	Folders   map[int]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Mailboxes: make(map[int]string),
		Folders:   make(map[int]string),
	}
}

type Thread struct {
	Type        string   `json:"type"`
	Action      string   `json:"type,omitempty"`
	CreatedBy   string   `json:"createdBy"`
	CreatedAt   string   `json:"createdAt"`
	Message     string   `json:"body,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

type Conversation struct {
	ID          int      `json:"id"`
	MailboxName string   `json:"mailboxName"`
	FolderName  string   `json:"folderName"`
	Subject     string   `json:"subject"`
	Customer    string   `json:"customer"`
	Threads     []Thread `json:"threads"`
	CreatedAt   string   `json:"createdAt"`
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}
}

func downloadAndSaveAttachment(storagePath string, url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	)
	req.Header.Set("Referer", "https://example.com")

	// Create an HTTP client and perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make the request: %v", err)
	}
	defer resp.Body.Close()

	var filename = strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
	var newSrc = storagePath + "/" + filename

	img, err := os.Create(newSrc)
	if err != nil {
		fmt.Println("Error creating image file:", err)
		return ""
	}

	_, err = io.Copy(img, resp.Body)
	if err != nil {
		fmt.Println("Error saving image:", err)
		return ""
	}

	return filename
}

func archiveConversation(c *Conversation) {
	fmt.Printf("Archiving conversation %s\n", c.Subject)

	var subject string
	if c.Subject == "" {
		subject = strings.Split(c.CreatedAt, "T")[0] + "_" + strconv.Itoa(c.ID) + "_" + c.Customer
	} else {
		trimmedSubject := c.Subject
		if len(trimmedSubject) > 52 {
			trimmedSubject = trimmedSubject[:52]
		}
		subject = strings.Split(c.CreatedAt, "T")[0] + "_" + strconv.Itoa(c.ID) + "_" + trimmedSubject
	}

	conversationStoragePath := "data/mailboxes/" + c.MailboxName + "/" + c.FolderName + "/" + subject
	createDirIfNotExist(conversationStoragePath)

	htmlContent := `
	<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title></title>
        <style>
            img { height: auto; width: 100%; margin: 0.25em; }
        </style>
    </head>
    <body>
    <h3>Subject: ` + c.Subject + `</h3>
    <p>Customer: ` + c.Customer + `</p>
    <p>Created at: ` + c.CreatedAt + `</p>
    <h3>Threads:</h3>
   	`

	for _, thread := range c.Threads {
		var attachments []string
		for _, attachment := range thread.Attachments {
			filename := downloadAndSaveAttachment(conversationStoragePath, attachment)
			if len(filename) > 0 {
				attachments = append(attachments, filename)
			}
		}

		if thread.Type == "lineitem" {
			htmlContent += `<div><h4 style="margin: 0; padding: 0">` + thread.CreatedAt + ` -> ` + thread.CreatedBy + `</h4>` + thread.Action + `</div>`
		} else {
			htmlContent += `<div><h4 style="margin: 0; paddinG: 0;">` + thread.CreatedAt + ` -> ` + thread.CreatedBy + `</h4>` + thread.Message + `</div>`
		}

		for _, attachment := range attachments {
			htmlContent += `<img src="` + attachment + `" style="max-width: 250px;"/>`
		}

		htmlContent += `<hr />`
	}

	htmlContent += `
    </body>
    </html>
	`

	file, err := os.Create(conversationStoragePath + "/index.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write the HTML content to the file
	_, err = file.WriteString(htmlContent)
	if err != nil {
		panic(err)
	}
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: go run main.go <argument>")
		return
	}

	storage := NewMemoryStorage()
	token := args[1]

	mb := mailbox.NewMailbox(token)

	mailboxes := mb.GetMailboxes(1)
	for _, mailbox := range mailboxes {
		storage.Mailboxes[mailbox.ID] = mailbox.Name

		folders := mb.GetFolders(mailbox.ID, 1)
		for _, folder := range folders {
			storage.Folders[folder.ID] = folder.Name
		}
	}

	conversations := mb.GetConversations(1)
	for _, conversation := range conversations {
		c := Conversation{
			ID:          conversation.ID,
			MailboxName: storage.Mailboxes[conversation.MailboxID],
			FolderName:  storage.Folders[conversation.FolderID],
			Subject:     conversation.Subject,
			Customer:    conversation.PrimaryCustomer.Email,
			CreatedAt:   conversation.CreatedAt,
		}

		for _, thread := range conversation.Embedded.Threads {
			t := Thread{
				Type:      thread.Type,
				Action:    thread.Action.Text,
				CreatedBy: thread.CreatedBy.Email,
				CreatedAt: thread.CreatedAt,
				Message:   thread.Body,
			}
			for _, attachment := range thread.Embedded.Attachments {
				t.Attachments = append(t.Attachments, attachment.Links.Web.Href)
			}
			c.Threads = append(c.Threads, t)
		}

		archiveConversation(&c)
	}
}
