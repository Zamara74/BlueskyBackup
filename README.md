# Bluesky Backup & Media Downloader (car-parser)

Ein natives Windows-GUI-Tool in Go zur Sicherung, Dekodierung und Aufbereitung von Bluesky `.car`-Archiven inklusive automatischer Medien-Downloads und verlinkter HTML-/TXT-Exports.

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