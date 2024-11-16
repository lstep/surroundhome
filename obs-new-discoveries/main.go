package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	authKey = "auth-key"
)

func initConfig() {
	// Set up command line flags
	pflag.String("obsidian-api-url", "http://localhost:27123", "Obsidian API URL")
	pflag.String(authKey, "", "Authentication key for Obsidian API")
	pflag.String("nats-address", "nats://localhost:4222", "NATS server address")
	pflag.Parse()

	// Bind flags to viper
	viper.BindPFlags(pflag.CommandLine)

	// Set up environment variables
	viper.SetEnvPrefix("OBS")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set up config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	fmt.Printf("Looking for config file in: %s\n", viper.ConfigFileUsed())

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		} else {
			fmt.Printf("Config file not found: %v\n", err)
		}
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Print current configuration
	fmt.Printf("Current configuration:\n")
	fmt.Printf("obsidian-api-url: %s\n", viper.GetString("obsidian-api-url"))
	fmt.Printf("nats-address: %s\n", viper.GetString("nats-address"))
	fmt.Printf("auth-key: %s\n", viper.GetString(authKey))

	// Set default values
	viper.SetDefault("obsidian-api-url", "http://localhost:27123")
	viper.SetDefault("nats-address", "nats://localhost:4222")
}

func main() {
	initConfig()

	nc, err := nats.Connect(viper.GetString("nats-address"))
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
	url := viper.GetString("obsidian-api-url") + "/periodic/daily/"

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
	req.Header.Add("Authorization", viper.GetString(authKey))
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
