package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const blueskyChatProxy = "did:web:api.bsky.chat#bsky_chat"

type ChatSender struct {
	DID string `json:"did"`
}

type ChatMessage struct {
	Sender ChatSender `json:"sender"`
	Text   string     `json:"text"`
	SentAt string     `json:"sentAt"`
}

type ChatMember struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
}

type ChatConversation struct {
	ID       string        `json:"id"`
	Members  []ChatMember  `json:"members"`
	Messages []ChatMessage `json:"messages"`
}

type chatSession struct {
	AccessJWT string `json:"accessJwt"`
	DID       string `json:"did"`
	Handle    string `json:"handle"`
}

type chatConvoList struct {
	Convos []ChatConversation `json:"convos"`
	Cursor string             `json:"cursor"`
}

type chatMessagesPage struct {
	Messages []ChatMessage `json:"messages"`
	Cursor   string        `json:"cursor"`
}

func FetchChatConversations(handle, appPassword, pdsHost string) ([]ChatConversation, string, error) {
	pdsURL, err := normalizePDSURL(pdsHost)
	if err != nil {
		return nil, "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	conversations, myDID, _, err := fetchChatConversations(context.Background(), client, pdsURL, handle, appPassword)
	return conversations, myDID, err
}

func normalizePDSURL(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("bitte gib den PDS-Server ein")
	}
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	parsed, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("ungültiger PDS-Server: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("der PDS-Server muss eine gültige HTTPS-Adresse ohne Pfad sein")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func inferPDSFromHandle(handle string) (string, bool) {
	handle = strings.TrimSpace(strings.TrimPrefix(handle, "@"))
	handle = strings.ToLower(strings.TrimSuffix(handle, "."))
	switch {
	case strings.HasSuffix(handle, ".bsky.social") && len(handle) > len(".bsky.social"):
		return "bsky.social", true
	case strings.HasSuffix(handle, ".eurosky.social") && len(handle) > len(".eurosky.social"):
		return "eurosky.social", true
	default:
		return "", false
	}
}

func fetchChatConversations(ctx context.Context, client *http.Client, pdsURL, handle, appPassword string) ([]ChatConversation, string, string, error) {
	handle = strings.TrimSpace(strings.TrimPrefix(handle, "@"))
	if handle == "" {
		return nil, "", "", fmt.Errorf("bitte gib dein Bluesky-Handle ein")
	}
	if strings.TrimSpace(appPassword) == "" {
		return nil, "", "", fmt.Errorf("bitte gib ein Bluesky-App-Passwort ein")
	}

	sessionBody, err := json.Marshal(map[string]string{
		"identifier": handle,
		"password":   appPassword,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("Anmeldedaten konnten nicht vorbereitet werden: %w", err)
	}
	var session chatSession
	if err := callChatXRPC(ctx, client, pdsURL, "com.atproto.server.createSession", nil, sessionBody, "", "", &session); err != nil {
		return nil, "", "", fmt.Errorf("Anmeldung bei Bluesky fehlgeschlagen: %w", err)
	}
	if session.AccessJWT == "" || session.DID == "" {
		return nil, "", "", fmt.Errorf("Bluesky hat keine gültige Sitzung zurückgegeben")
	}

	conversations, err := listChatConversations(ctx, client, pdsURL, session.AccessJWT)
	if err != nil {
		return nil, "", "", err
	}
	for i := range conversations {
		messages, err := listConversationMessages(ctx, client, pdsURL, session.AccessJWT, conversations[i].ID)
		if err != nil {
			return nil, "", "", fmt.Errorf("Nachrichten einer Unterhaltung konnten nicht geladen werden: %w", err)
		}
		conversations[i].Messages = messages
	}
	return conversations, session.DID, session.Handle, nil
}

func callChatXRPC(ctx context.Context, client *http.Client, baseURL, method string, params url.Values, body []byte, accessJWT, proxyService string, result any) error {
	endpoint, err := url.Parse(strings.TrimRight(baseURL, "/") + "/xrpc/" + method)
	if err != nil {
		return fmt.Errorf("ungültige XRPC-Adresse: %w", err)
	}
	if params != nil {
		endpoint.RawQuery = params.Encode()
	}

	var requestBody io.Reader
	requestMethod := http.MethodGet
	if body != nil {
		requestBody = bytes.NewReader(body)
		requestMethod = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, requestMethod, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("XRPC-Anfrage konnte nicht erstellt werden: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessJWT != "" {
		req.Header.Set("Authorization", "Bearer "+accessJWT)
	}
	if proxyService != "" {
		req.Header.Set("atproto-proxy", proxyService)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Netzwerkfehler bei %s: %w", method, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return fmt.Errorf("%s antwortete mit HTTP %d (Fehlerdetails konnten nicht gelesen werden: %w)", method, resp.StatusCode, readErr)
		}
		details := strings.TrimSpace(string(responseBody))
		if details == "" {
			return fmt.Errorf("%s antwortete mit HTTP %d", method, resp.StatusCode)
		}
		return fmt.Errorf("%s antwortete mit HTTP %d: %s", method, resp.StatusCode, details)
	}
	if result == nil {
		_, err := io.Copy(io.Discard, resp.Body)
		return err
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("Antwort von %s konnte nicht gelesen werden: %w", method, err)
	}
	return nil
}

func callChatServiceXRPC(ctx context.Context, client *http.Client, pdsURL, method string, params url.Values, accessJWT string, result any) error {
	return callChatXRPC(ctx, client, pdsURL, method, params, nil, accessJWT, blueskyChatProxy, result)
}

func listChatConversations(ctx context.Context, client *http.Client, pdsURL, accessJWT string) ([]ChatConversation, error) {
	var conversations []ChatConversation
	cursor := ""
	seenCursors := make(map[string]struct{})
	for {
		params := url.Values{"limit": {"100"}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var page chatConvoList
		if err := callChatServiceXRPC(ctx, client, pdsURL, "chat.bsky.convo.listConvos", params, accessJWT, &page); err != nil {
			return nil, fmt.Errorf("Unterhaltungen konnten nicht geladen werden: %w", err)
		}
		conversations = append(conversations, page.Convos...)
		if page.Cursor == "" {
			return conversations, nil
		}
		if _, exists := seenCursors[page.Cursor]; exists {
			return nil, fmt.Errorf("Bluesky hat bei Unterhaltungen einen wiederholten Seiten-Cursor zurückgegeben")
		}
		seenCursors[page.Cursor] = struct{}{}
		cursor = page.Cursor
	}
}

func listConversationMessages(ctx context.Context, client *http.Client, pdsURL, accessJWT, convoID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	cursor := ""
	seenCursors := make(map[string]struct{})
	for {
		params := url.Values{
			"convoId": {convoID},
			"limit":   {"100"},
		}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var page chatMessagesPage
		if err := callChatServiceXRPC(ctx, client, pdsURL, "chat.bsky.convo.getMessages", params, accessJWT, &page); err != nil {
			return nil, fmt.Errorf("chat.bsky.convo.getMessages: %w", err)
		}
		messages = append(messages, page.Messages...)
		if page.Cursor == "" {
			sort.SliceStable(messages, func(i, j int) bool {
				left, leftErr := time.Parse(time.RFC3339Nano, messages[i].SentAt)
				right, rightErr := time.Parse(time.RFC3339Nano, messages[j].SentAt)
				if leftErr == nil && rightErr != nil {
					return true
				}
				if leftErr != nil && rightErr == nil {
					return false
				}
				if leftErr != nil {
					return messages[i].SentAt < messages[j].SentAt
				}
				return left.Before(right)
			})
			return messages, nil
		}
		if _, exists := seenCursors[page.Cursor]; exists {
			return nil, fmt.Errorf("Bluesky hat bei Nachrichten einen wiederholten Seiten-Cursor zurückgegeben")
		}
		seenCursors[page.Cursor] = struct{}{}
		cursor = page.Cursor
	}
}

func ExportChatsToHTML(conversations []ChatConversation, targetFile, myDID string) error {
	if len(conversations) == 0 {
		return fmt.Errorf("es wurden keine Chats gefunden")
	}

	sort.SliceStable(conversations, func(i, j int) bool {
		return strings.ToLower(chatPartnerName(conversations[i].Members, myDID)) <
			strings.ToLower(chatPartnerName(conversations[j].Members, myDID))
	})

	var document strings.Builder
	document.WriteString(`<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Bluesky-Chats</title>
<style>
  body { font-family: sans-serif; background: #f0f2f5; padding: 20px; color: #17202a; }
  h1, .overview, .conversation-list { max-width: 900px; margin-left: auto; margin-right: auto; }
  h1 { margin-top: 0; }
  .overview, .conversation { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .overview { margin-bottom: 24px; }
  .overview ul { list-style: none; padding: 0; margin: 0; }
  .overview li + li { border-top: 1px solid #e4e6eb; }
  .overview-link { display: flex; justify-content: space-between; gap: 16px; padding: 12px 4px; color: #075eaa; text-decoration: none; }
  .overview-link:hover, .overview-link:focus { text-decoration: underline; }
  .overview-meta { color: #667085; text-align: right; white-space: nowrap; }
  .conversation { margin-bottom: 16px; overflow: auto; }
  .conversation summary { cursor: pointer; font-weight: 600; }
  .conversation summary::marker { color: #0084ff; }
  .conversation-meta { color: #667085; font-size: 0.9em; font-weight: normal; margin-left: 8px; }
  .msg { margin-bottom: 10px; padding: 8px 12px; border-radius: 12px; max-width: 75%; clear: both; overflow-wrap: anywhere; }
  .sent { background: #0084ff; color: white; float: right; }
  .received { background: #e4e6eb; color: black; float: left; }
  .time { font-size: 0.75em; opacity: 0.7; margin-top: 4px; display: block; text-align: right; }
  @media (max-width: 600px) {
    body { padding: 12px; }
    .overview-link { flex-direction: column; gap: 4px; }
    .overview-meta { text-align: left; white-space: normal; }
    .msg { max-width: 90%; }
  }
</style>
</head>
<body>
<h1>Bluesky-Chats</h1>
<section class="overview">
<h2>Übersicht</h2>
<p>` + fmt.Sprintf("%d Unterhaltungen", len(conversations)) + ` – wähle einen Chat, um ihn zu öffnen.</p>
<ul>
`)

	for i, conversation := range conversations {
		partnerName := chatPartnerName(conversation.Members, myDID)
		lastActivity := chatLastActivity(conversation.Messages)
		chatID := fmt.Sprintf("chat-%d", i+1)
		document.WriteString(fmt.Sprintf(`<li><a class="overview-link" href="#%s" data-chat-link><span>%s</span><span class="overview-meta">%s · zuletzt %s</span></a></li>`+"\n",
			chatID, html.EscapeString(partnerName), chatMessageCountLabel(len(conversation.Messages)), html.EscapeString(lastActivity)))
	}
	document.WriteString(`</ul>
</section>
<main class="conversation-list">
`)

	for i, conversation := range conversations {
		partnerName := chatPartnerName(conversation.Members, myDID)
		chatID := fmt.Sprintf("chat-%d", i+1)
		document.WriteString(fmt.Sprintf(`<details class="conversation" id="%s"><summary>%s<span class="conversation-meta">%s · zuletzt %s</span></summary>`+"\n",
			chatID, html.EscapeString(partnerName), chatMessageCountLabel(len(conversation.Messages)), html.EscapeString(chatLastActivity(conversation.Messages))))
		for _, message := range conversation.Messages {
			class := "received"
			if message.Sender.DID == myDID {
				class = "sent"
			}
			formattedTime := message.SentAt
			if parsedTime, err := time.Parse(time.RFC3339Nano, message.SentAt); err == nil {
				formattedTime = parsedTime.Local().Format("02.01.2006 15:04")
			}
			document.WriteString(fmt.Sprintf(`<div class="msg %s">%s<span class="time">%s</span></div>`+"\n",
				class, html.EscapeString(message.Text), html.EscapeString(formattedTime)))
		}
		document.WriteString("</details>\n")
	}
	document.WriteString(`</main>
<script>
document.querySelectorAll("[data-chat-link]").forEach(function (link) {
  link.addEventListener("click", function () {
    var chat = document.getElementById(link.getAttribute("href").slice(1));
    if (chat) chat.open = true;
  });
});
</script>
</body>
</html>
`)

	if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
		return fmt.Errorf("Zielordner konnte nicht erstellt werden: %w", err)
	}
	if err := os.WriteFile(targetFile, []byte(document.String()), 0644); err != nil {
		return fmt.Errorf("HTML-Datei konnte nicht gespeichert werden: %w", err)
	}
	return nil
}

func chatMessageCountLabel(count int) string {
	if count == 1 {
		return "1 Nachricht"
	}
	return fmt.Sprintf("%d Nachrichten", count)
}

func chatLastActivity(messages []ChatMessage) string {
	var latest time.Time
	latestValue := ""
	for _, message := range messages {
		parsed, err := time.Parse(time.RFC3339Nano, message.SentAt)
		if err != nil {
			if latestValue == "" {
				latestValue = message.SentAt
			}
			continue
		}
		if latest.IsZero() || parsed.After(latest) {
			latest = parsed
		}
	}
	if !latest.IsZero() {
		return latest.Local().Format("02.01.2006 15:04")
	}
	if latestValue != "" {
		return latestValue
	}
	return "keine Nachrichten"
}

func chatPartnerName(members []ChatMember, myDID string) string {
	var names []string
	for _, member := range members {
		if member.DID == myDID {
			continue
		}
		name := member.Handle
		if name != "" {
			name = "@" + strings.TrimPrefix(name, "@")
		} else if member.DisplayName != "" {
			name = member.DisplayName
		} else {
			name = member.DID
		}
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "Unbekannter Chatpartner"
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
