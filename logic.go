package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	bsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

type OutputRecord struct {
	Cid  string      `json:"cid"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type MediaTask struct {
	URL  string
	Path string
}

func nodeToInterface(n interface{}) interface{} {
	node, ok := n.(ipld.Node)
	if !ok || node == nil {
		return n
	}
	switch node.Kind() {
	case ipld.Kind_Map:
		m := make(map[string]interface{})
		itr := node.MapIterator()
		for !itr.Done() {
			k, v, _ := itr.Next()
			kStr, _ := k.AsString()
			m[kStr] = nodeToInterface(v)
		}
		return m
	case ipld.Kind_List:
		var list []interface{}
		itr := node.ListIterator()
		for !itr.Done() {
			_, v, _ := itr.Next()
			list = append(list, nodeToInterface(v))
		}
		return list
	case ipld.Kind_String:
		s, _ := node.AsString()
		return s
	case ipld.Kind_Bool:
		b, _ := node.AsBool()
		return b
	case ipld.Kind_Int:
		i, _ := node.AsInt()
		return i
	case ipld.Kind_Float:
		f, _ := node.AsFloat()
		return f
	case ipld.Kind_Bytes:
		b, _ := node.AsBytes()
		return b
	case ipld.Kind_Link:
		link, _ := node.AsLink()
		if link != nil {
			return map[string]interface{}{
				"$link": link.String(),
			}
		}
		return nil
	default:
		return nil
	}
}

func parseCarFile(carPath string, baseOutputDir string, logFunc func(string)) error {
	outputDir := filepath.Join(baseOutputDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	f, err := os.Open(carPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader, err := car.NewBlockReader(f)
	if err != nil {
		return err
	}

	recordsByType := make(map[string][]OutputRecord)
	seenCids := make(map[string]bool)

	logFunc("Lese CAR-Blöcke ein...")

	for {
		block, err := reader.Next()
		if err != nil {
			break
		}

		blockCid := block.Cid().String()
		if seenCids[blockCid] {
			continue
		}

		nb := basicnode.Prototype.Any.NewBuilder()
		if err := dagcbor.Decode(nb, bytes.NewReader(block.RawData())); err != nil {
			continue
		}
		node := nb.Build()

		typeField, err := node.LookupByString("$type")
		if err != nil {
			continue
		}
		typeStr, err := typeField.AsString()
		if err != nil || typeStr == "" {
			continue
		}

		seenCids[blockCid] = true

		var recData interface{}

		if typeStr == "app.bsky.feed.post" {
			var post bsky.FeedPost
			if errCBOR := post.UnmarshalCBOR(bytes.NewReader(block.RawData())); errCBOR == nil {
				recData = &post
			}
		}

		if recData == nil {
			recData = nodeToInterface(node)
		}

		recordsByType[typeStr] = append(recordsByType[typeStr], OutputRecord{
			Cid:  blockCid,
			Type: typeStr,
			Data: recData,
		})
	}

	for typeStr, records := range recordsByType {
		if typeStr == "app.bsky.feed.post" {
			sort.Slice(records, func(i, j int) bool {
				p1, ok1 := records[i].Data.(*bsky.FeedPost)
				p2, ok2 := records[j].Data.(*bsky.FeedPost)
				if ok1 && ok2 {
					return p1.CreatedAt < p2.CreatedAt
				}
				return false
			})
		}

		safeTypeName := strings.ReplaceAll(typeStr, ".", "_")
		filePath := filepath.Join(outputDir, fmt.Sprintf("output_%s.json", safeTypeName))

		jsonBytes, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			continue
		}

		_ = os.WriteFile(filePath, jsonBytes, 0644)
		logFunc(fmt.Sprintf("Exportiert: output_%s.json (%d Einträge)", safeTypeName, len(records)))
	}

	return nil
}

func downloadMedia(baseOutputDir, userDid, pdsHost string, progressFunc func(current, total int, msg string)) error {
	postFile := filepath.Join(baseOutputDir, "output", "output_app_bsky_feed_post.json")
	medienOrdner := filepath.Join(baseOutputDir, "Medien")
	_ = os.MkdirAll(medienOrdner, 0755)

	data, err := os.ReadFile(postFile)
	if err != nil {
		return fmt.Errorf("Post-JSON nicht gefunden: %v", err)
	}

	var records []OutputRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("Fehler beim Lesen der Posts: %v", err)
	}

	var tasks []MediaTask

	for _, rec := range records {
		postDataBytes, err := json.Marshal(rec.Data)
		if err != nil {
			continue
		}

		var post bsky.FeedPost
		if err := json.Unmarshal(postDataBytes, &post); err != nil || post.Embed == nil || post.Embed.EmbedImages == nil {
			continue
		}

		datumClean := "Unbekannt"
		if post.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, post.CreatedAt); err == nil {
				datumClean = t.Format("2006-01-02_15-04-05")
			} else if t, err := time.Parse(time.RFC3339, post.CreatedAt); err == nil {
				datumClean = t.Format("2006-01-02_15-04-05")
			}
		}

		for idx, img := range post.Embed.EmbedImages.Images {
			if img.Image == nil {
				continue
			}
			blobCid := img.Image.Ref.String()
			if blobCid != "" {
				fileName := fmt.Sprintf("%s_%d.jpg", datumClean, idx+1)
				downloadUrl := fmt.Sprintf("https://%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s", pdsHost, userDid, blobCid)
				tasks = append(tasks, MediaTask{
					URL:  downloadUrl,
					Path: filepath.Join(medienOrdner, fileName),
				})
			}
		}
	}

	total := len(tasks)
	if total == 0 {
		progressFunc(0, 0, "Keine Bilder zum Herunterladen gefunden.")
		return nil
	}

	progressFunc(0, total, fmt.Sprintf("Starte Download von %d Bildern...", total))

	var wg sync.WaitGroup
	guard := make(chan struct{}, 8)
	var completed int
	var mu sync.Mutex

	client := &http.Client{Timeout: 30 * time.Second}

	for _, task := range tasks {
		if _, err := os.Stat(task.Path); err == nil {
			mu.Lock()
			completed++
			progressFunc(completed, total, fmt.Sprintf("Bereits vorhanden: %s", filepath.Base(task.Path)))
			mu.Unlock()
			continue
		}

		wg.Add(1)
		guard <- struct{}{}

		go func(t MediaTask) {
			defer wg.Done()
			defer func() { <-guard }()

			resp, err := client.Get(t.URL)
			if err == nil && resp.StatusCode == 200 {
				out, err := os.Create(t.Path)
				if err == nil {
					_, _ = io.Copy(out, resp.Body)
					out.Close()
				}
				resp.Body.Close()
			}

			mu.Lock()
			completed++
			progressFunc(completed, total, fmt.Sprintf("Geladen: %s", filepath.Base(t.Path)))
			mu.Unlock()
		}(task)
	}

	wg.Wait()
	return nil
}

// Export 1: Posts (Neueste zuerst + Bildhinweise mit Datums-Dateinamen)
func exportPostsAsCleanText(baseOutputDir, targetFilePath string) error {
	postFile := filepath.Join(baseOutputDir, "output", "output_app_bsky_feed_post.json")
	if _, err := os.Stat(postFile); os.IsNotExist(err) {
		return fmt.Errorf("Keine Post-JSON gefunden! Bitte zuerst 'Backup starten' ausführen.")
	}

	data, err := os.ReadFile(postFile)
	if err != nil {
		return err
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("============================================================\n")
	sb.WriteString("BLUESKY BEITRÄGE EXPORT\n")
	sb.WriteString(fmt.Sprintf("Exportiert am: %s\n", time.Now().Format("02.01.2006 15:04:05")))
	sb.WriteString(fmt.Sprintf("Anzahl Beiträge: %d\n", len(records)))
	sb.WriteString("============================================================\n\n")

	total := len(records)
	for i := total - 1; i >= 0; i-- {
		item := records[i]
		recordData, ok := item["data"].(map[string]interface{})
		if !ok {
			continue
		}

		text, _ := recordData["text"].(string)
		createdAt, _ := recordData["createdAt"].(string)
		postNum := i + 1

		// Deutsches Datumsformat generieren
		formattedDate := formatGermanDateTime(createdAt)

		sb.WriteString(fmt.Sprintf("[%d] %s\n", postNum, formattedDate))
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		sb.WriteString(text + "\n")
		
		// --- Bildreferenzen mit Datums-Dateinamen auslesen ---
		if embed, ok := recordData["embed"].(map[string]interface{}); ok {
			var images []interface{}

			if imgs, ok := embed["images"].([]interface{}); ok {
				images = imgs
			} else if media, ok := embed["media"].(map[string]interface{}); ok {
				if imgs, ok := media["images"].([]interface{}); ok {
					images = imgs
				}
			}

			if len(images) > 0 {
				sb.WriteString("\nAngehängte Bilder:\n")

				// Erstellungsdatum aus 'createdAt' für den Dateinamen parsen (z. B. 2026-10-25T19:15:22.423Z -> 2026-10-25_19-15-22)
				datePrefix := "bild"
				if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
					datePrefix = parsedTime.Format("2006-01-02_15-04-05")
				} else if len(createdAt) >= 19 {
					// Fallback: Manuelles Aufräumen des ISO-Strings, falls Parsing fehlschlägt
					datePrefix = strings.ReplaceAll(createdAt[:19], "T", "_")
					datePrefix = strings.ReplaceAll(datePrefix, ":", "-")
				}

				for imgIndex := range images {
					// Bildnummerierung passend zum Speicher-Schema (z.B. 2026-10-25_19-15-22_1.jpeg)
					imgFilename := fmt.Sprintf("%s_%d.jpeg", datePrefix, imgIndex+1)
					sb.WriteString(fmt.Sprintf(" - [Bild: %s]\n", imgFilename))
				}
			}
		}

		sb.WriteString("\n\n")
	}

	return os.WriteFile(targetFilePath, []byte(sb.String()), 0644)
}
// Export: Follows mit Handle-Auflösung
func exportFollowsAsCleanText(baseOutputDir, targetFilePath string) error {
	file := filepath.Join(baseOutputDir, "output", "output_app_bsky_graph_follow.json")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("Keine Follows-JSON gefunden! Bitte zuerst 'Backup starten' ausführen.")
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("============================================================\n")
	sb.WriteString(fmt.Sprintf("GEFOLGTE KONTEN (%d Einträge)\n", len(records)))
	sb.WriteString(fmt.Sprintf("Exportiert am: %s\n", time.Now().Format("02.01.2006 15:04:05")))
	sb.WriteString("============================================================\n\n")

	for i, item := range records {
		recordData, ok := item["data"].(map[string]interface{})
		if !ok {
			continue
		}
		subjectDID, _ := recordData["subject"].(string)
		createdAt, _ := recordData["createdAt"].(string)

		// Versuch, DID in Handle umzuwandeln (sonst Fallback auf DID)
		handleOrDID := resolveDIDToHandle(subjectDID, "bsky.social")

		sb.WriteString(fmt.Sprintf("[%d] [%s] %s\n", i+1, createdAt, handleOrDID))
	}

	return os.WriteFile(targetFilePath, []byte(sb.String()), 0644)
}

// Export: Blocks mit Handle-Auflösung
func exportBlocksAsCleanText(baseOutputDir, targetFilePath string) error {
	file := filepath.Join(baseOutputDir, "output", "output_app_bsky_graph_block.json")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("Keine Block-JSON gefunden! Bitte zuerst 'Backup starten' ausführen.")
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("============================================================\n")
	sb.WriteString(fmt.Sprintf("BLOCKIERTE KONTEN (%d Einträge)\n", len(records)))
	sb.WriteString(fmt.Sprintf("Exportiert am: %s\n", time.Now().Format("02.01.2006 15:04:05")))
	sb.WriteString("============================================================\n\n")

	for i, item := range records {
		recordData, ok := item["data"].(map[string]interface{})
		if !ok {
			continue
		}
		subjectDID, _ := recordData["subject"].(string)
		createdAt, _ := recordData["createdAt"].(string)

		// Versuch, DID in Handle umzuwandeln (sonst Fallback auf DID)
		handleOrDID := resolveDIDToHandle(subjectDID, "bsky.social")

		sb.WriteString(fmt.Sprintf("[%d] [%s] %s\n", i+1, createdAt, handleOrDID))
	}
	return os.WriteFile(targetFilePath, []byte(sb.String()), 0644)
}

// Export 4: Likes
func exportLikesAsCleanText(baseOutputDir, targetFilePath string) error {
	file := filepath.Join(baseOutputDir, "output", "output_app_bsky_feed_like.json")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("Keine Likes-JSON gefunden!")
	}

	data, _ := os.ReadFile(file)
	var records []map[string]interface{}
	json.Unmarshal(data, &records)

	var sb strings.Builder
	sb.WriteString("============================================================\n")
	sb.WriteString(fmt.Sprintf("GELIKTE BEITRÄGE (Geliked: %d)\n", len(records)))
	sb.WriteString("============================================================\n\n")

	for _, item := range records {
		recordData, ok := item["data"].(map[string]interface{})
		if !ok {
			continue
		}
		subject, _ := recordData["subject"].(map[string]interface{})
		uri, _ := subject["uri"].(string)
		createdAt, _ := recordData["createdAt"].(string)

		sb.WriteString(fmt.Sprintf("[%s] %s\n", createdAt, uri))
	}

	return os.WriteFile(targetFilePath, []byte(sb.String()), 0644)
}
func resolveHandleToDID(handle, pdsHost string) (string, error) {
	if strings.HasPrefix(handle, "did:") {
		return handle, nil
	}

	cleanHandle := strings.TrimSpace(handle)
	cleanHandle = strings.TrimPrefix(cleanHandle, "@")

	url := fmt.Sprintf("https://%s/xrpc/com.atproto.identity.resolveHandle?handle=%s", pdsHost, cleanHandle)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("Netzwerkfehler bei der DID-Auflösung: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Handle '%s' konnte nicht gefunden werden (Status %d)", cleanHandle, resp.StatusCode)
	}

	var result struct {
		Did string `json:"did"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("Antwort konnte nicht gelesen werden: %v", err)
	}

	return result.Did, nil
}
// Cache, damit gleiche DIDs nicht mehrfach online abgefragt werden
var didHandleCache = make(map[string]string)

// Wandelt eine DID in einen Handle um (z. B. did:plc:... -> @username.bsky.social).
// Bei Fehlern oder ungültigen DIDs wird als Fallback die originale DID zurückgegeben.
func resolveDIDToHandle(did, pdsHost string) string {
	cleanDID := strings.TrimSpace(did)

	// Falls es gar keine DID ist oder der Cache den Eintrag schon kennt
	if !strings.HasPrefix(cleanDID, "did:") {
		return cleanDID
	}
	if cachedHandle, found := didHandleCache[cleanDID]; found {
		return cachedHandle
	}

	// API-Aufruf zur DID-Auflösung über PLC Directory
	url := fmt.Sprintf("https://plc.directory/%s", cleanDID)
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		didHandleCache[cleanDID] = cleanDID // Fallback speichern
		return cleanDID
	}
	defer resp.Body.Close()

	var result struct {
		AlsoKnownAs []string `json:"alsoKnownAs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.AlsoKnownAs) == 0 {
		didHandleCache[cleanDID] = cleanDID // Fallback speichern
		return cleanDID
	}

	// 'at://username.bsky.social' zu '@username.bsky.social' umwandeln
	handle := strings.TrimPrefix(result.AlsoKnownAs[0], "at://")
	formattedHandle := "@" + handle

	didHandleCache[cleanDID] = formattedHandle
	return formattedHandle
}
// Export 1b: Posts als HTML-Dokument mit anklickbaren Bild-Links
func exportPostsAsHTML(baseOutputDir, targetFilePath string, hasDownloadedMedia bool) error {
	postFile := filepath.Join(baseOutputDir, "output", "output_app_bsky_feed_post.json")
	if _, err := os.Stat(postFile); os.IsNotExist(err) {
		return fmt.Errorf("Keine Post-JSON gefunden! Bitte zuerst 'Backup starten' ausführen.")
	}

	data, err := os.ReadFile(postFile)
	if err != nil {
		return err
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"de\">\n<head>\n")
	sb.WriteString("<meta charset=\"UTF-8\">\n")
	sb.WriteString("<title>Bluesky Beiträge Export</title>\n")
	sb.WriteString("<style>\n")
	sb.WriteString("  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; max-width: 800px; margin: 20px auto; padding: 0 20px; background-color: #f9f9f9; color: #222; }\n")
	sb.WriteString("  h1 { border-bottom: 2px solid #0085ff; padding-bottom: 10px; color: #0085ff; }\n")
	sb.WriteString("  .post { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 15px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.05); }\n")
	sb.WriteString("  .post-header { font-size: 0.85em; color: #666; margin-bottom: 10px; border-bottom: 1px solid #eee; padding-bottom: 5px; }\n")
	sb.WriteString("  .post-text { font-size: 1em; white-space: pre-wrap; word-wrap: break-word; }\n")
	sb.WriteString("  .media-box { margin-top: 12px; padding: 10px; background: #f0f7ff; border-radius: 6px; border-left: 4px solid #0085ff; font-size: 0.9em; }\n")
	sb.WriteString("  .media-link { color: #0085ff; text-decoration: none; font-weight: bold; }\n")
	sb.WriteString("  .media-link:hover { text-decoration: underline; }\n")
	sb.WriteString("</style>\n</head>\n<body>\n")

	sb.WriteString("<h1>Bluesky Beiträge Export</h1>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Exportiert am:</strong> %s | <strong>Anzahl Beiträge:</strong> %d</p><hr>\n\n", time.Now().Format("02.01.2006 15:04:05"), len(records)))

	medienDir := filepath.Join(baseOutputDir, "Medien")
	mediaFolderExists := false
	if _, err := os.Stat(medienDir); err == nil {
		mediaFolderExists = true
	}

	total := len(records)
	for i := total - 1; i >= 0; i-- {
		item := records[i]
		recordData, ok := item["data"].(map[string]interface{})
		if !ok {
			continue
		}

		text, _ := recordData["text"].(string)
		createdAt, _ := recordData["createdAt"].(string)
		postNum := i + 1

		// Deutsches Datumsformat generieren
		formattedDate := formatGermanDateTime(createdAt)

		sb.WriteString("<div class=\"post\">\n")
		sb.WriteString(fmt.Sprintf("  <div class=\"post-header\">#%d &bull; %s</div>\n", postNum, formattedDate))
		sb.WriteString(fmt.Sprintf("  <div class=\"post-text\">%s</div>\n", strings.ReplaceAll(text, "\n", "<br>")))

		// --- Bildverarbeitung ---
		if embed, ok := recordData["embed"].(map[string]interface{}); ok {
			var images []interface{}
			if imgs, ok := embed["images"].([]interface{}); ok {
				images = imgs
			} else if media, ok := embed["media"].(map[string]interface{}); ok {
				if imgs, ok := media["images"].([]interface{}); ok {
					images = imgs
				}
			}

			if len(images) > 0 {
				sb.WriteString("  <div class=\"media-box\">\n    <strong>Angehängte Bilder:</strong><br>\n")

				datePrefix := "bild"
				if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
					datePrefix = parsedTime.Format("2006-01-02_15-04-05")
				} else if len(createdAt) >= 19 {
					datePrefix = strings.ReplaceAll(createdAt[:19], "T", "_")
					datePrefix = strings.ReplaceAll(datePrefix, ":", "-")
				}

				for imgIndex := range images {
    // Geändert von .jpeg auf .jpg
    imgFilename := fmt.Sprintf("%s_%d.jpg", datePrefix, imgIndex+1)
    relPath := fmt.Sprintf("Medien/%s", imgFilename)

    if hasDownloadedMedia && mediaFolderExists {
        // Anklickbarer Link zur lokalen Datei
        sb.WriteString(fmt.Sprintf("    &bull; <a class=\"media-link\" href=\"%s\" target=\"_blank\">Bild öffnen: %s</a><br>\n", relPath, imgFilename))
    } else {
        // Nur Text-Hinweis ohne Hyperlink
        sb.WriteString(fmt.Sprintf("    &bull; [Bild: %s] <em>(nicht heruntergeladen)</em><br>\n", imgFilename))
    }
}
				sb.WriteString("  </div>\n")
			}
		}

		sb.WriteString("</div>\n\n")
	}

	sb.WriteString("</body>\n</html>")

	return os.WriteFile(targetFilePath, []byte(sb.String()), 0644)
}
// Formatiert einen ISO-Timestamp (z.B. 2026-10-28T19:35:24.022Z) in deutsches Format
func formatGermanDateTime(isoString string) string {
	// ISO-8601 Strings parsen (UTC)
	t, err := time.Parse(time.RFC3339, isoString)
	if err != nil {
		// Fallback, falls Parsing scheitert
		if len(isoString) >= 19 {
			return strings.ReplaceAll(isoString[:10], "-", ".") + " - " + isoString[11:19] + " Uhr"
		}
		return isoString
	}

	// In lokale Zeit umwandeln oder in UTC im deutschen Format ausgeben (28.10.2026 - 19:35:24 Uhr)
	return t.Format("02.01.2006 - 15:04:05") + " Uhr"
}