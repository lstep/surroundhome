# OBS API Documentation

This service listens for NATS messages to add new discoveries/links to your Obsidian daily notes.

## Base Configuration
- NATS Server: `nats://localhost:4222`
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

### Testing from the RESTProxy API interface
The RESTProxy API interface can be used to test the service. The service is listening on port 8080 by default.

To test the service, you can use the following cURL command:
```bash
curl -X POST -H "Content-Type: application/json" -d '{"content": "https://example.com"}' http://localhost:8080/memorize
```

or you can use httpie:
```bash
http POST http://localhost:8080/memorize content="https://example.com"
```

# TODO
- Implement an optional generation of the description of the content of the URL using an LLM after converting the
page to markdown (https://github.com/JohannesKaufmann/html-to-markdown for example)
