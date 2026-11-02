<img width="602" height="573" alt="Screenshot 2026-10-30 161557" src="https://github.com/user-attachments/assets/e262936c-6fb1-43f2-9097-f04920766095" />

English | [Deutsch](README.de.md)

# Bluesky Backup & Media Downloader (car-parser)

**Author:** [Zamara74](https://github.com/Zamara74)

A native Windows GUI tool written in Go for backing up, decoding, and preparing Bluesky `.car` archives, including automatic media downloads and linked HTML/TXT exports.

## 🚀 Getting started

### 1. Download your CAR file from Bluesky
1. Open **Bluesky** in your browser (Settings → Account → *Export account data* / *Konto-Daten exportieren*).
2. Download your repository archive (the `.car` file, e.g. `repo.car`).

### 2. Run the tool and back up your data
1. Start `BlueskyBackup.exe`.
2. Select the `.car` file you downloaded.
3. Enter your **user handle** (e.g. `yourname.bsky.social`). The PDS server is detected automatically for handles under `bsky.social` and `eurosky.social`. For custom handle domains or a DID, enter the PDS server manually in the **PDS Server** field.
4. To export DMs, create an app password in Bluesky under **Settings → Privacy and security → App passwords**. It is not stored and is only used to retrieve your messages.
5. Choose an **output folder** and click **Process CAR file & download media**.
6. Then click the export buttons at the bottom. Posts can be saved as a neatly formatted `.html` file (with image links) or as `.txt`. **Chats / DMs** are retrieved directly from Bluesky and saved as `chats_lesbar.html` in the output folder. The file includes an overview with the message count and latest activity; clicking a chat partner opens the corresponding conversation.

## Features
- **Complete parsing:** Processes `.car` archives locally (posts, follows, blocks, likes).
- **Media downloads:** Downloads embedded images via the AT Protocol API and names them chronologically.
- **Chat export:** After signing in with an app password, retrieves DMs directly from Bluesky and saves them in a single HTML file, sorted by chat partner, with a clickable overview and collapsible conversations.
- **Clean exports:**
  - Posts as `.html` with clickable image links (in the relative `Medien/` folder), or as plain `.txt`.
  - Resolves DIDs to handles via `plc.directory` with in-memory caching (falls back to the DID for inactive accounts).
  - Chronological ordering (newest entries first, oldest last with `#1`).
- **German and English:** The user interface automatically follows the system language. It appears in German when the system language is German and in English otherwise. Both translations are embedded in the application; no additional language files are required.

## Windows release build
The Windows build requires **Go 1.23.5 or later** and **w64devkit** with the GCC C compiler. Fyne uses CGO for the native Windows GUI, so `CGO_ENABLED` must be set to `1` and `gcc.exe` must be available on `PATH`.

Install w64devkit and adjust the path below to match its installation location. Run these commands in a PowerShell terminal in the project folder:

```powershell
$env:PATH += ";C:\w64devkit\bin"
$env:CGO_ENABLED = "1"
fyne package -os windows -icon icon.png -name BlueskyBackup
```

If the build fails, first check whether the compiler can be found by running `gcc --version`. The environment variables apply only to the current PowerShell window.

The command uses `icon.png` for the EXE file icon. The window and taskbar icon is also loaded from `icon.png`.
The language strings are embedded in the program, so the build can still be distributed as a single EXE without external language files.

## Libraries and open-source services
This project builds on the following open-source projects and APIs:

* **[Fyne](https://fyne.io/)** (BSD 3-Clause License) – Cross-platform GUI toolkit for Go.
* **[Go-ATProto / CAR Parsing](https://github.com/bluesky-social/indigo)** – AT Protocol specifications and tools.
* **[PLC Directory](https://plc.directory/)** – Public DID resolution for Bluesky identities.

## Credits and acknowledgements
* The original logic for parsing `.car` files is based on the open-source script [`car-to-json`](https://github.com/romiojoseph/bluesky/tree/main/go-scripts/car-to-json) by **[romiojoseph](https://github.com/romiojoseph)** (MIT License).
* This project extends the script with a graphical user interface (Fyne), native Windows dialogs, media downloads via the AT Protocol, DID resolution, and HTML/TXT export features.

## License
This project is distributed under the [MIT License](LICENSE).
