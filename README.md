# Bluesky Backup & Media Downloader (car-parser)

**Autor:** [Zamara74](https://github.com/Zamara74)

Ein natives Windows-GUI-Tool in Go zur Sicherung, Dekodierung und Aufbereitung von Bluesky `.car`-Archiven inklusive automatischer Medien-Downloads und verlinkter HTML-/TXT-Exports.

## 🚀 Anleitung & Nutzung

### 1. CAR-Datei von Bluesky herunterladen
1. Öffne **Bluesky** im Browser (Einstellungen -> Konto -> *Konto-Daten exportieren* / *Export Data*).
2. Lade dein Repository-Archiv (die `.car`-Datei, z. B. `repo.car`) herunter.

### 2. Tool starten & Daten sichern
1. Starte die `BlueskyBackup.exe`.
2. Wähle deine heruntergeladene `.car`-Datei aus.
3. Gib dein **User-Handle** (z. B. `deinname.bsky.social`) ein. Für Handles unter `bsky.social` und `eurosky.social` wird der PDS-Server automatisch erkannt. Bei eigenen Handle-Domains oder einer DID trägst du den PDS-Server im Feld **PDS-Server** manuell ein.
4. Für den DM-Export erstellst du in Bluesky unter **Einstellungen → Datenschutz und Sicherheit → App-Passwörter** ein App-Passwort. Dieses wird nicht gespeichert und nur für den Abruf verwendet.
5. Wähle einen **Zielordner** und klicke auf **1. CAR-Backup & Medien herunterladen**.
6. Klicke anschließend im unteren Bereich auf die Export-Buttons. Posts lassen sich als übersichtliche `.html` (mit Bild-Links) oder `.txt` speichern. **Chats / DMs** werden direkt von Bluesky geladen und als `chats_lesbar.html` im Zielordner abgelegt. Die Datei enthält eine Übersicht mit Nachrichtenzahl und letzter Aktivität; ein Klick auf einen Chatpartner öffnet den zugehörigen Chatverlauf.

## Features
- **Nativer Dateidialog:** Nutzt Windows-Shell-Dialoge ohne Konsolenfenster oder UI-Freezes.
- **Vollständiges Parsing:** Verarbeitet `.car`-Archive lokal (Posts, Follows, Blocks, Likes).
- **Medien-Download:** Lädt eingebettete Bilder über die AT-Protocol-API herunter und benennt sie chronologisch.
- **Chat-Export:** Lädt DMs nach Anmeldung mit einem App-Passwort direkt von Bluesky und speichert sie gesammelt in einer nach Chatpartner sortierten HTML-Datei mit klickbarer Übersicht und einklappbaren Verläufen.
- **Saubere Exporte:**
  - Posts als `.html` mit anklickbaren Bild-Links (im relativen Ordner `Medien/`) oder als reine `.txt`.
  - Auflösung von DIDs zu Handles über `plc.directory` mit In-Memory-Caching (Fallback auf DID bei inaktiven Accounts).
  - Chronologische Sortierung (neueste Einträge oben, älteste unten mit `#1`).

## Release-Build für Windows
Im PowerShell-Terminal im Projektordner:

```powershell
$env:PATH += ";E:\Programme\w64devkit\bin"
$env:CGO_ENABLED = "1"
fyne package -os windows -icon icon.png -name BlueskyBackup
```

Der Befehl verwendet `icon.png` für das EXE-Dateisymbol. Das Fenster-/Taskleisten-Icon wird ebenfalls aus `icon.png` geladen.

## Verwendete Bibliotheken & Open-Source-Dienste
Dieses Projekt baut auf folgenden Open-Source-Projekten und APIs auf:

* **[Fyne](https://fyne.io/)** (BSD 3-Clause License) – Cross-Platform GUI Toolkit für Go.
* **[Go-ATProto / CAR Parsing](https://github.com/bluesky-social/indigo)** – AT Protocol Spezifikation & Tools.
* **[PLC Directory](https://plc.directory/)** – Öffentliche DID-Auflösung für Bluesky-Identitäten.

## Credits & Danksagung
* Die ursprüngliche Logik zum Parsen der `.car`-Dateien basiert auf dem Open-Source-Skript [`car-to-json`](https://github.com/romiojoseph/bluesky/tree/main/go-scripts/car-to-json) von **[romiojoseph](https://github.com/romiojoseph)** (MIT-Lizenz).
* Das Skript wurde in diesem Projekt um eine grafische Benutzeroberfläche (Fyne), native Windows-Dialoge, Medien-Downloads über das AT-Protokoll, DID-Auflösung sowie HTML-/TXT-Exportfunktionen erweitert.

## Lizenz
Dieses Projekt steht unter der [MIT License](LICENSE).