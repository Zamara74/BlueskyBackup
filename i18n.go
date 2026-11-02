package main

import (
	"log"
	"strings"

	"github.com/jeandeaual/go-locale"
)

type uiText struct {
	windowTitle              string
	carPathPlaceholder       string
	browseFiles              string
	handlePlaceholder        string
	appPasswordPlaceholder   string
	pdsPlaceholder           string
	browseFolder             string
	startWithMedia           string
	startWithoutMedia        string
	downloadImages           string
	statusReady              string
	exportAsText             string
	exportAsHTMLFile         string
	exportAsHTML             string
	errorTitle               string
	missingHandle            string
	missingPassword          string
	pdsRequiredTitle         string
	pdsRequiredMessage       string
	loginDownloadingChats    string
	chatDownloadError        string
	chatHTMLSaveError        string
	chatsSaved               string
	openSaveDialog           string
	textFileFilter           string
	htmlFileFilter           string
	allFilesFilter           string
	saveCancelled            string
	creatingTextFile         string
	savedSuccessfully        string
	openHTMLSaveDialog       string
	creatingHTML             string
	htmlSavedSuccessfully    string
	missingCAR               string
	handleRequiredForMedia   string
	pdsNotRecognized         string
	startingCARParsing       string
	parsingError             string
	resolvingHandle          string
	handleResolutionError    string
	didFoundStartingDownload string
	downloadError            string
	backupComplete           string
	exportComplete           string
	carFileLabel             string
	handleDIDLabel           string
	appPasswordLabel         string
	pdsLabel                 string
	outputFolderLabel        string
	optionsLabel             string
	postsLabel               string
	followsLabel             string
	blocksLabel              string
	likesLabel               string
	chatsLabel               string
	cleanDataLabel           string
	carFileFilter            string
	selectCarTitle           string
	selectFolderDescription  string
	saveFileTitle            string
	readingCARBlocks         string
	exportedRecords          string
	noImagesFound            string
	startingImageDownload    string
	imageAlreadyExists       string
	imageDownloaded          string
}

type uiLanguage string

const (
	german  uiLanguage = "de"
	english uiLanguage = "en"
)

func detectUILanguage() uiLanguage {
	language, err := locale.GetLanguage()
	if err != nil {
		log.Printf("Could not detect system language, using English: %v", err)
		return english
	}

	return languageForTag(language)
}

func languageForTag(tag string) uiLanguage {
	language, _, _ := strings.Cut(tag, "-")
	if strings.EqualFold(language, string(german)) {
		return german
	}
	return english
}

func textsFor(language uiLanguage) uiText {
	if language == german {
		return uiText{
			windowTitle:              "Bluesky Backup & Media Downloader",
			carPathPlaceholder:       "Pfad zur .car-Datei...",
			browseFiles:              "Durchsuchen...",
			handlePlaceholder:        "z.B. username.bsky.social oder did:plc:...",
			appPasswordPlaceholder:   "Bluesky-App-Passwort",
			pdsPlaceholder:           "Wird erkannt oder manuell eingeben",
			browseFolder:             "Ordner wählen...",
			startWithMedia:           "CAR-Datei verarbeiten & Medien herunterladen",
			startWithoutMedia:        "CAR-Datei verarbeiten",
			downloadImages:           "Bilder herunterladen",
			statusReady:              "Bereit.",
			exportAsText:             "Als TXT...",
			exportAsHTMLFile:         "Als HTML...",
			exportAsHTML:             "HTML (online)...",
			errorTitle:               "Fehler",
			missingHandle:            "Bitte gib dein Bluesky-Handle ein.",
			missingPassword:          "Bitte gib ein Bluesky-App-Passwort ein.",
			pdsRequiredTitle:         "PDS-Server erforderlich",
			pdsRequiredMessage:       "Für dieses Handle kann der PDS-Server nicht automatisch erkannt werden. Bitte gib ihn manuell ein.",
			loginDownloadingChats:    "Melde bei Bluesky an und lade Chats...",
			chatDownloadError:        "Fehler beim Chat-Download: %v",
			chatHTMLSaveError:        "Fehler beim Speichern der Chat-HTML: %v",
			chatsSaved:               "Chats als HTML gespeichert: %s",
			openSaveDialog:           "Öffne Speichern-Dialog...",
			textFileFilter:           "Textdateien (*.txt)|*.txt|Alle Dateien (*.*)|*.*",
			htmlFileFilter:           "HTML-Dateien (*.html)|*.html|Alle Dateien (*.*)|*.*",
			allFilesFilter:           "Alle Dateien (*.*)|*.*",
			saveCancelled:            "Speichern abgebrochen.",
			creatingTextFile:         "Erstelle Textdatei...",
			savedSuccessfully:        "Erfolgreich gespeichert unter: %s",
			openHTMLSaveDialog:       "Öffne Speichern-Dialog für HTML...",
			creatingHTML:             "Erstelle HTML-Dokument...",
			htmlSavedSuccessfully:    "HTML erfolgreich gespeichert: %s",
			missingCAR:               "Bitte wähle eine .car-Datei aus!",
			handleRequiredForMedia:   "Für den Bilder-Download wird ein Handle oder eine DID benötigt!",
			pdsNotRecognized:         "Der PDS-Server konnte für dieses Handle nicht automatisch erkannt werden. Bitte gib ihn manuell ein.",
			startingCARParsing:       "Starte CAR-Parsing...",
			parsingError:             "Fehler beim Parsing: %v",
			resolvingHandle:          "Löse Handle/DID auf...",
			handleResolutionError:    "Fehler bei Handle-Auflösung: %v",
			didFoundStartingDownload: "DID gefunden (%s). Starte Medien-Download...",
			downloadError:            "Download-Fehler: %v",
			backupComplete:           "Fertig! Alle JSONs und Medien wurden erfolgreich gespeichert.",
			exportComplete:           "Fertig! JSON-Dateien wurden erfolgreich exportiert (Bilder übersprungen).",
			carFileLabel:             "CAR-Datei:",
			handleDIDLabel:           "User Handle / DID:",
			appPasswordLabel:         "App-Passwort für DMs:",
			pdsLabel:                 "PDS-Server:",
			outputFolderLabel:        "Zielordner:",
			optionsLabel:             "Optionen:",
			postsLabel:               "Posts",
			followsLabel:             "Follows",
			blocksLabel:              "Blocks",
			likesLabel:               "Likes",
			chatsLabel:               "Chats / DMs",
			cleanDataLabel:           "Daten bereinigen:",
			carFileFilter:            "CAR-Dateien (*.car)|*.car|Alle Dateien (*.*)|*.*",
			selectCarTitle:           ".car-Datei auswählen",
			selectFolderDescription:  "Zielordner für den Export auswählen",
			saveFileTitle:            "Datei speichern unter",
			readingCARBlocks:         "Lese CAR-Blöcke ein...",
			exportedRecords:          "Exportiert: output_%s.json (%d Einträge)",
			noImagesFound:            "Keine Bilder zum Herunterladen gefunden.",
			startingImageDownload:    "Starte Download von %d Bildern...",
			imageAlreadyExists:       "Bereits vorhanden: %s",
			imageDownloaded:          "Geladen: %s",
		}
	}

	return uiText{
		windowTitle:              "Bluesky Backup & Media Downloader",
		carPathPlaceholder:       "Path to .car file...",
		browseFiles:              "Browse...",
		handlePlaceholder:        "e.g. username.bsky.social or did:plc:...",
		appPasswordPlaceholder:   "Bluesky app password",
		pdsPlaceholder:           "Detected automatically or enter manually",
		browseFolder:             "Choose folder...",
		startWithMedia:           "Process CAR file & download media",
		startWithoutMedia:        "Process CAR file",
		downloadImages:           "Download images",
		statusReady:              "Ready.",
		exportAsText:             "As TXT...",
		exportAsHTMLFile:         "As HTML...",
		exportAsHTML:             "HTML (online)...",
		errorTitle:               "Error",
		missingHandle:            "Please enter your Bluesky handle.",
		missingPassword:          "Please enter a Bluesky app password.",
		pdsRequiredTitle:         "PDS server required",
		pdsRequiredMessage:       "The PDS server for this handle could not be detected automatically. Please enter it manually.",
		loginDownloadingChats:    "Signing in to Bluesky and downloading chats...",
		chatDownloadError:        "Chat download failed: %v",
		chatHTMLSaveError:        "Failed to save chat HTML: %v",
		chatsSaved:               "Chats saved as HTML: %s",
		openSaveDialog:           "Opening save dialog...",
		textFileFilter:           "Text files (*.txt)|*.txt|All files (*.*)|*.*",
		htmlFileFilter:           "HTML files (*.html)|*.html|All files (*.*)|*.*",
		allFilesFilter:           "All files (*.*)|*.*",
		saveCancelled:            "Saving cancelled.",
		creatingTextFile:         "Creating text file...",
		savedSuccessfully:        "Successfully saved to: %s",
		openHTMLSaveDialog:       "Opening save dialog for HTML...",
		creatingHTML:             "Creating HTML document...",
		htmlSavedSuccessfully:    "HTML saved successfully: %s",
		missingCAR:               "Please select a .car file!",
		handleRequiredForMedia:   "A handle or DID is required to download images!",
		pdsNotRecognized:         "The PDS server for this handle could not be detected automatically. Please enter it manually.",
		startingCARParsing:       "Starting CAR parsing...",
		parsingError:             "Parsing failed: %v",
		resolvingHandle:          "Resolving handle/DID...",
		handleResolutionError:    "Handle resolution failed: %v",
		didFoundStartingDownload: "Found DID (%s). Starting media download...",
		downloadError:            "Download failed: %v",
		backupComplete:           "Done! All JSON files and media were saved successfully.",
		exportComplete:           "Done! JSON files were exported successfully (images skipped).",
		carFileLabel:             "CAR file:",
		handleDIDLabel:           "User handle / DID:",
		appPasswordLabel:         "App password for DMs:",
		pdsLabel:                 "PDS server:",
		outputFolderLabel:        "Output folder:",
		optionsLabel:             "Options:",
		postsLabel:               "Posts",
		followsLabel:             "Follows",
		blocksLabel:              "Blocks",
		likesLabel:               "Likes",
		chatsLabel:               "Chats / DMs",
		cleanDataLabel:           "Clean up data:",
		carFileFilter:            "CAR files (*.car)|*.car|All files (*.*)|*.*",
		selectCarTitle:           "Select a .car file",
		selectFolderDescription:  "Select an output folder",
		saveFileTitle:            "Save file as",
		readingCARBlocks:         "Reading CAR blocks...",
		exportedRecords:          "Exported: output_%s.json (%d records)",
		noImagesFound:            "No images to download were found.",
		startingImageDownload:    "Starting download of %d images...",
		imageAlreadyExists:       "Already exists: %s",
		imageDownloaded:          "Downloaded: %s",
	}
}
