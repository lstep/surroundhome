package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nats-io/nats.go"
)

const (
	obsidianAPIURL = "http://localhost:27123"
	authKey        = "Bearer 40800f7c8a05343d8fdd8c5d610d2f3fa95d58c72a6b68fe52bbb209ff90bfd5"
)

func main() {
	nc, err := nats.Connect("nats://ailocal:4222")
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	fmt.Println("Connected to NATS, running...")

	nc.Subscribe("POST:.memorize", func(msg *nats.Msg) {
		fmt.Printf("Received: %#v\n", msg.Subject)

		fmt.Printf("Received message data: %s\n", string(msg.Data))

		// Retrieve the content from the message
		var data struct {
			URL      string   `json:"url"`
			Title    string   `json:"title"`
			Tags     []string `json:"tags"`
			Selected string   `json:"selected"`
		}

		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Printf("Error decoding JSON: %v\n", err)
			return
		}

		if data.URL == "" {
			fmt.Printf("Empty url!\n")
			msg.Respond([]byte("Error: URL is empty"))
			return
		}

		url := data.URL

		fmt.Printf("URL: %s title: %s tags: %v text: %s\n", url, data.Title, data.Tags, data.Selected)

		// Publish to Obsidian Daily Note
		err = publishToObsidianDailyNote(url, data.Tags, data.Selected)
		if err != nil {
			fmt.Printf("Error publishing to Obsidian: %v\n", err)
		}

		msg.Respond([]byte("ACK"))
	})

	select {}
}

func publishToObsidianDailyNote(content string, tags []string, selected string) error {
	// Get today's date for the Daily Note
	//today := time.Now().Format("2006-01-02")

	// Construct the API request
	url := fmt.Sprintf("%s/periodic/daily/", obsidianAPIURL)

	// build content
	title, err := getTitleFromURL(content)
	if err != nil {
		fmt.Printf("Failed to get title for %s: %#v\n", content, err)
		title = content
	}
	content = fmt.Sprintf("- NewDiscovery:: [%s](%s)", title, content)
	if len(tags) > 0 {
		content += " #" + strings.Join(tags, " #")
	}
	if len(selected) > 0 {
		content += "\n     - " + selected
	}

	payload := strings.NewReader(content)

	req, err := http.NewRequest(http.MethodPatch, url, payload)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Add("Content-Type", "text/markdown")
	req.Header.Add("Authorization", authKey)
	req.Header.Add("Content-Insertion-Position", "end")
	req.Header.Add("Heading", "::New discoveries")
	req.Header.Add("Target", "::New discoveries")
	req.Header.Add("Operation", "append")
	req.Header.Add("Target-Type", "heading")
	req.Header.Add("Target-Delimiter", "::")
	req.Header.Add("Trim-Target-Whitespace", "false")
	req.Header.Add("Target-Delimiter", "::")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%#v\n", resp.Status)
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("client: response body: %s\n", resBody)

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Println("Successfully published to Obsidian Daily Note")
	return nil
}
