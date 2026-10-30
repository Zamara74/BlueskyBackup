<img width="680" height="574" alt="Screenshot 2026-10-30 112909" src="https://github.com/user-attachments/assets/9ae817b0-763f-498b-b7ea-28bdee6ee5e1" />

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
3. Gib dein **User-Handle** (z. B. `deinname.bsky.social`) ein, damit das Tool deine DID und Medien zuordnen kann.
4. Wähle einen **Zielordner** und klicke auf **1. CAR-Backup & Medien herunterladen**.
5. Klicke anschließend im unteren Bereich auf die **Bereinigen-Buttons**, um deine Posts als übersichtliche `.html` (mit Bild-Links) oder `.txt` zu speichern.

## Features
- **Nativer Dateidialog:** Nutzt Windows-Shell-Dialoge ohne Konsolenfenster oder UI-Freezes.
- **Vollständiges Parsing:** Verarbeitet `.car`-Archive lokal (Posts, Follows, Blocks, Likes).
- **Medien-Download:** Lädt eingebettete Bilder über die AT-Protocol-API herunter und benennt sie chronologisch.
- **Saubere Exporte:**
  - Posts als `.html` mit anklickbaren Bild-Links (im relativen Ordner `Medien/`) oder als reine `.txt`.
  - Auflösung von DIDs zu Handles über `plc.directory` mit In-Memory-Caching (Fallback auf DID bei inaktiven Accounts).
  - Chronologische Sortierung (neueste Einträge oben, älteste unten mit `#1`).

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
