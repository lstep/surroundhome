package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	slog.Info("looking for config file", "path", viper.ConfigFileUsed())

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("error reading config file", "error", err)
		} else {
			slog.Info("config file not found", "error", err)
		}
	} else {
		slog.Info("using config file", "path", viper.ConfigFileUsed())
	}

	// Print current configuration
	slog.Info("current configuration",
		"obsidian-api-url", viper.GetString("obsidian-api-url"),
		"nats-address", viper.GetString("nats-address"),
		"auth-key", viper.GetString(authKey))

	// Set default values
	viper.SetDefault("obsidian-api-url", "http://localhost:27123")
	viper.SetDefault("nats-address", "nats://localhost:4222")
}

func main() {
	// Configure slog with JSON handler and debug level
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	initConfig()

	nc, err := nats.Connect(viper.GetString("nats-address"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	slog.Info("connected to NATS, running...")

	nc.Subscribe("memorize", handleMemorizeMessage)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigChan
	slog.Info("received signal, initiating graceful shutdown", "signal", sig)

	// Perform cleanup
	if err := nc.Drain(); err != nil {
		slog.Error("error draining NATS connection", "error", err)
	}
}

func handleMemorizeMessage(msg *nats.Msg) {
	slog.Info("received message", "subject", msg.Subject)
	slog.Debug("message data", "data", string(msg.Data))

	// Retrieve the content from the message
	var data struct {
		URL      string   `json:"url"`
		Title    string   `json:"title"`
		Tags     []string `json:"tags"`
		Selected string   `json:"selected"`
	}

	err := json.Unmarshal(msg.Data, &data)
	if err != nil {
		slog.Error("error decoding JSON", "error", err)
		return
	}

	if data.URL == "" {
		slog.Error("empty url received")
		msg.Respond([]byte("Error: URL is empty"))
		return
	}

	url := data.URL

	slog.Info("processing message",
		"url", url,
		"title", data.Title,
		"tags", data.Tags,
		"selected_text", data.Selected)

	// Publish to Obsidian Daily Note
	slog.Info("publishing to Obsidian Daily Note")
	err = publishToObsidianDailyNote(url, data.Tags, data.Selected)
	if err != nil {
		slog.Error("error publishing to Obsidian", "error", err)
		errResp, _ := json.Marshal(map[string]string{"error": "error publishing to Obsidian"})
		msg.Respond(errResp)
		return
	}

	slog.Debug("sending ACK")
	resp, _ := json.Marshal(map[string]string{"error": ""})
	msg.Respond(resp)
}

func publishToObsidianDailyNote(content string, tags []string, selected string) error {
	// Create HTTP client with recommended parameters
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false,
		},
	}

	// Get today's date for the Daily Note
	//today := time.Now().Format("2006-01-02")

	// Construct the API request
	url := viper.GetString("obsidian-api-url") + "/periodic/daily/"

	// build content
	title, err := getTitleFromURL(content)
	if err != nil {
		slog.Error("failed to get title for URL", "url", content, "error", err)
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

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected response", "status", resp.Status)
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("could not read response body", "error", err)
			os.Exit(1)
		}
		slog.Error("response body", "body", string(resBody))

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	slog.Info("successfully published to Obsidian Daily Note")
	return nil
}
