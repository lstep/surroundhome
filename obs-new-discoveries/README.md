# New Discoveries API Documentation

This service listens for NATS messages to add new discoveries/links to your Obsidian daily notes.

## Base Configuration
- NATS Server: `nats://ailocal:4222`
- Subscription Topic: `POST:.memorize`

## Adding a New Discovery

### Endpoint
Send a NATS message to `POST:.memorize`

### Request Format
```json
{
    "content": "https://example.com"
}
```

### Parameters
- `content` (string, required): The URL or content you want to add to your daily notes

### Response
- Success: `"ACK"`
- Error: Error message string (e.g., `"Error: Content is empty"`)

### Example Usage

Using NATS CLI:
```bash
nats pub "POST:.memorize" '{"content": "https://example.com"}'
```

Using Go:
```go
nc, _ := nats.Connect("nats://ailocal:4222")
payload := `{"content": "https://example.com"}`
msg, _ := nc.Request("POST:.memorize", []byte(payload), time.Second)
```

### Behavior
1. The service will attempt to fetch the title of the URL (if it's a URL)
2. It will format the content as a new discovery entry
3. The entry will be added to your Obsidian daily note under the "::New discoveries" heading
4. The format in Obsidian will be: `- NewDiscovery:: [Title](URL)`

### Error Cases
- Empty content will return an error
- Invalid JSON format will result in an error
- Failed connection to Obsidian API will result in an error

Note: The service requires a running Obsidian instance with the Local REST API plugin configured with the appropriate authentication key.
