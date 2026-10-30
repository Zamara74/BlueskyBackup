package main

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.png
var appIcon []byte

func main() {
	myApp := app.New()
	myApp.SetIcon(fyne.NewStaticResource("icon.png", appIcon))
	myWindow := myApp.NewWindow("Bluesky Backup & Media Downloader")
	myWindow.Resize(fyne.NewSize(680, 620))
	creditsLabel := widget.NewLabelWithStyle("v1.1.0 by Zamara74", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	carPathEntry := widget.NewEntry()
	carPathEntry.SetPlaceHolder("Pfad zur .car-Datei...")

	btnBrowseCar := widget.NewButton("Durchsuchen...", func() {
		go func() {
			if selectedFile, ok := openNativeOpenFileDialog(); ok {
				fyne.Do(func() {
					carPathEntry.SetText(selectedFile)
				})
			}
		}()
	})

	handleEntry := widget.NewEntry()
	handleEntry.SetPlaceHolder("z.B. username.bsky.social oder did:plc:...")

	chatPasswordEntry := widget.NewPasswordEntry()
	chatPasswordEntry.SetPlaceHolder("Bluesky-App-Passwort")

	pdsEntry := widget.NewEntry()
	pdsEntry.SetPlaceHolder("Wird erkannt oder manuell eingeben")
	pdsAutoDetected := false
	handleEntry.OnChanged = func(handle string) {
		if pds, ok := inferPDSFromHandle(handle); ok {
			pdsAutoDetected = true
			pdsEntry.SetText(pds)
			pdsEntry.Disable()
			return
		}
		if pdsAutoDetected {
			pdsEntry.SetText("")
			pdsAutoDetected = false
		}
		pdsEntry.Enable()
	}

	outPathEntry := widget.NewEntry()
	outPathEntry.SetText(".")

	btnBrowseOut := widget.NewButton("Ordner wählen...", func() {
		go func() {
			if selectedFolder, ok := openNativeFolderDialog(); ok {
				fyne.Do(func() {
					outPathEntry.SetText(selectedFolder)
				})
			}
		}()
	})

	// 1. Button zuerst mit Standard-Text und HighImportance erstellen
	btnStart := widget.NewButton("1. CAR-Datei verarbeiten & Medien herunterladen", func() {
		// Hier kommt deine bisherige Start-Funktion/Logik rein
	})
	btnStart.Importance = widget.HighImportance

	// 2. Eine EINZIGE Checkbox erstellen, die Text-Wechsel UND Haken steuert
	chkDownloadMedia := widget.NewCheck("Bilder herunterladen", func(checked bool) {
		if checked {
			btnStart.SetText("CAR-Datei verarbeiten & Medien herunterladen")
		} else {
			btnStart.SetText("CAR-Datei verarbeiten")
		}
	})
	chkDownloadMedia.SetChecked(true) // Standardmäßig aktiviert

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)

	statusLabel := widget.NewLabel("Bereit.")
	statusLabel.Wrapping = fyne.TextWrapWord
	setStatus := func(message string) {
		fyne.Do(func() {
			statusLabel.SetText(message)
		})
	}

	// --- EXPORT-BEREICH ---
	btnCleanPostsTXT := widget.NewButton("Als TXT...", nil)
	btnCleanPostsHTML := widget.NewButton("Als HTML...", nil)

	btnCleanFollows := widget.NewButton("Als TXT...", nil)
	btnCleanBlocks := widget.NewButton("Als TXT...", nil)
	btnCleanLikes := widget.NewButton("Als TXT...", nil)

	// Button für den Online-Chat-Export
	backupRunning := false
	var btnCleanChats *widget.Button
	btnCleanChats = widget.NewButton("HTML (online)...", func() {
		if handleEntry.Text == "" {
			dialog.ShowInformation("Fehler", "Bitte gib dein Bluesky-Handle ein.", myWindow)
			return
		}
		if chatPasswordEntry.Text == "" {
			dialog.ShowInformation("Fehler", "Bitte gib ein Bluesky-App-Passwort ein.", myWindow)
			return
		}
		if pdsEntry.Text == "" {
			dialog.ShowInformation("PDS-Server erforderlich", "Für dieses Handle kann der PDS-Server nicht automatisch erkannt werden. Bitte gib ihn manuell ein.", myWindow)
			return
		}

		handle := handleEntry.Text
		appPassword := chatPasswordEntry.Text
		pds := pdsEntry.Text
		outDir := outPathEntry.Text
		if outDir == "" {
			outDir = "."
		}
		chatPasswordEntry.SetText("")
		btnCleanChats.Disable()
		statusLabel.SetText("Melde bei Bluesky an und lade Chats...")

		go func() {
			conversations, myDID, err := FetchChatConversations(handle, appPassword, pds)
			appPassword = ""
			if err != nil {
				setStatus(fmt.Sprintf("Fehler beim Chat-Download: %v", err))
				fyne.Do(func() {
					if !backupRunning {
						btnCleanChats.Enable()
					}
				})
				return
			}

			targetPath := filepath.Join(outDir, "chats_lesbar.html")
			err = ExportChatsToHTML(conversations, targetPath, myDID)
			if err != nil {
				setStatus(fmt.Sprintf("Fehler beim Speichern der Chat-HTML: %v", err))
				fyne.Do(func() {
					if !backupRunning {
						btnCleanChats.Enable()
					}
				})
				return
			}

			setStatus(fmt.Sprintf("Chats als HTML gespeichert: %s", targetPath))
			fyne.Do(func() {
				if !backupRunning {
					btnCleanChats.Enable()
				}
			})
		}()
	})

	setExportButtonsState := func(enabled bool) {
		fyne.Do(func() {
			backupRunning = !enabled
			if enabled {
				btnCleanPostsTXT.Enable()
				btnCleanPostsHTML.Enable()
				btnCleanFollows.Enable()
				btnCleanBlocks.Enable()
				btnCleanLikes.Enable()
				btnCleanChats.Enable()
			} else {
				btnCleanPostsTXT.Disable()
				btnCleanPostsHTML.Disable()
				btnCleanFollows.Disable()
				btnCleanBlocks.Disable()
				btnCleanLikes.Disable()
				btnCleanChats.Disable()
			}
		})
	}

	// Event-Handler für TXT-Exports
	handleExport := func(defaultName string, exportFunc func(string, string) error) {
		outDir := outPathEntry.Text
		go func() {
			setStatus("Öffne Speichern-Dialog...")
			filter := "Textdateien (*.txt)|*.txt|Alle Dateien (*.*)|*.*"

			targetPath, ok := openNativeSaveDialog(defaultName, filter)
			if !ok || targetPath == "" {
				setStatus("Speichern abgebrochen.")
				return
			}

			setStatus("Erstelle Textdatei...")
			err := exportFunc(outDir, targetPath)
			if err != nil {
				setStatus(fmt.Sprintf("Fehler: %v", err))
			} else {
				setStatus(fmt.Sprintf("Erfolgreich gespeichert unter: %s", filepath.Base(targetPath)))
			}
		}()
	}

	btnCleanPostsTXT.OnTapped = func() { handleExport("posts_lesbar.txt", exportPostsAsCleanText) }

	btnCleanPostsHTML.OnTapped = func() {
		outDir := outPathEntry.Text
		isMediaChecked := chkDownloadMedia.Checked

		go func() {
			setStatus("Öffne Speichern-Dialog für HTML...")
			filter := "HTML-Dateien (*.html)|*.html|Alle Dateien (*.*)|*.*"

			targetPath, ok := openNativeSaveDialog("posts_lesbar.html", filter)
			if !ok || targetPath == "" {
				setStatus("Speichern abgebrochen.")
				return
			}

			setStatus("Erstelle HTML-Dokument...")
			err := exportPostsAsHTML(outDir, targetPath, isMediaChecked)
			if err != nil {
				setStatus(fmt.Sprintf("Fehler: %v", err))
			} else {
				setStatus(fmt.Sprintf("HTML erfolgreich gespeichert unter: %s", filepath.Base(targetPath)))
			}
		}()
	}

	btnCleanFollows.OnTapped = func() { handleExport("follows_lesbar.txt", exportFollowsAsCleanText) }
	btnCleanBlocks.OnTapped = func() { handleExport("blocks_lesbar.txt", exportBlocksAsCleanText) }
	btnCleanLikes.OnTapped = func() { handleExport("likes_lesbar.txt", exportLikesAsCleanText) }

	// Backup Start-Button
	btnStart.OnTapped = func() {
		carPath := carPathEntry.Text
		handleInput := handleEntry.Text
		pds := pdsEntry.Text
		outDir := outPathEntry.Text
		downloadMediaChecked := chkDownloadMedia.Checked

		if carPath == "" {
			dialog.ShowError(fmt.Errorf("Bitte wähle eine .car-Datei aus!"), myWindow)
			return
		}

		if downloadMediaChecked && handleInput == "" {
			dialog.ShowError(fmt.Errorf("Für den Bilder-Download wird ein Handle oder eine DID benötigt!"), myWindow)
			return
		}
		if downloadMediaChecked && pdsEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Der PDS-Server konnte für dieses Handle nicht automatisch erkannt werden. Bitte gib ihn manuell ein."), myWindow)
			return
		}

		btnStart.Disable()
		backupRunning = true
		setExportButtonsState(false)
		progressBar.SetValue(0)

		go func() {
			setStatus("Starte CAR-Parsing...")
			err := parseCarFile(carPath, outDir, func(msg string) {
				setStatus(msg)
			})

			if err != nil {
				setStatus(fmt.Sprintf("Fehler beim Parsing: %v", err))
				fyne.Do(func() {
					btnStart.Enable()
				})
				setExportButtonsState(true)
				return
			}

			if downloadMediaChecked {
				setStatus("Löse Handle/DID auf...")
				userDid, err := resolveHandleToDID(handleInput, pds)
				if err != nil {
					setStatus(fmt.Sprintf("Fehler bei Handle-Auflösung: %v", err))
					fyne.Do(func() {
						btnStart.Enable()
					})
					setExportButtonsState(true)
					return
				}

				setStatus(fmt.Sprintf("DID gefunden (%s). Starte Medien-Download...", userDid))

				err = downloadMedia(outDir, userDid, pds, func(current, total int, msg string) {
					setStatus(msg)
					if total > 0 {
						value := float64(current) / float64(total)
						fyne.Do(func() {
							progressBar.SetValue(value)
						})
					}
				})

				if err != nil {
					setStatus(fmt.Sprintf("Download-Fehler: %v", err))
				} else {
					setStatus("Fertig! Alle JSONs und Medien wurden erfolgreich gespeichert.")
				}
			} else {
				fyne.Do(func() {
					progressBar.SetValue(1.0)
				})
				setStatus("Fertig! JSON-Dateien wurden erfolgreich exportiert (Bilder übersprungen).")
			}

			fyne.Do(func() {
				btnStart.Enable()
			})
			setExportButtonsState(true)
		}()
	}

	carRow := container.NewBorder(nil, nil, nil, btnBrowseCar, carPathEntry)
	outRow := container.NewBorder(nil, nil, nil, btnBrowseOut, outPathEntry)

	form := widget.NewForm(
		widget.NewFormItem("CAR-Datei:", carRow),
		widget.NewFormItem("User Handle / DID:", handleEntry),
		widget.NewFormItem("App-Passwort für DMs:", chatPasswordEntry),
		widget.NewFormItem("PDS-Server:", pdsEntry),
		widget.NewFormItem("Zielordner:", outRow),
		widget.NewFormItem("Optionen:", chkDownloadMedia),
	)

	postsButtons := container.NewHBox(btnCleanPostsTXT, btnCleanPostsHTML)

	leftExportColumn := container.New(layout.NewFormLayout(),
		widget.NewLabel("Posts"), postsButtons,
		widget.NewLabel("Follows"), container.NewHBox(btnCleanFollows),
		widget.NewLabel("Blocks"), container.NewHBox(btnCleanBlocks),
	)
	rightExportColumn := container.New(layout.NewFormLayout(),
		widget.NewLabel("Likes"), container.NewHBox(btnCleanLikes),
		widget.NewLabel("Chats / DMs"), container.NewHBox(btnCleanChats),
	)
	exportGrid := container.NewGridWithColumns(2, leftExportColumn, rightExportColumn)

	// Statusleiste: Links Status, Rechts Credits
	bottomBar := container.NewBorder(nil, nil, nil, creditsLabel, statusLabel)

	// Gesamter Inhalt in einer vertikalen Box
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		btnStart,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Daten bereinigen:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		exportGrid,
		widget.NewSeparator(),
		progressBar,
		bottomBar,
	)

	// Padded-Container sorgt für einen gleichmäßigen Innenabstand zum Fensterrahmen
	myWindow.SetContent(container.NewPadded(content))

	// Fenstergröße automatisch an den Inhalt anpassen (verhindert das unnötige Aufblähen)
	myWindow.Resize(fyne.NewSize(600, 0))

	myWindow.SetContent(container.NewPadded(content))
	myWindow.ShowAndRun()
}
