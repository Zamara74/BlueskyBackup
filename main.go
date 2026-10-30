package main

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Bluesky Backup & Media Downloader")
	myWindow.Resize(fyne.NewSize(680, 620))
	creditsLabel := widget.NewLabelWithStyle("v1.0.0 by Zamara74", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	carPathEntry := widget.NewEntry()
	carPathEntry.SetPlaceHolder("Pfad zur .car-Datei...")

	btnBrowseCar := widget.NewButton("Durchsuchen...", func() {
		go func() {
			if selectedFile, ok := openNativeOpenFileDialog(); ok {
				carPathEntry.SetText(selectedFile)
			}
		}()
	})

	handleEntry := widget.NewEntry()
	handleEntry.SetPlaceHolder("z.B. username.bsky.social oder did:plc:...")

	pdsEntry := widget.NewEntry()
	pdsEntry.SetText("bsky.social")

	outPathEntry := widget.NewEntry()
	outPathEntry.SetText(".")

	btnBrowseOut := widget.NewButton("Ordner wählen...", func() {
		go func() {
			if selectedFolder, ok := openNativeFolderDialog(); ok {
				outPathEntry.SetText(selectedFolder)
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
			btnStart.SetText("1. CAR-Datei verarbeiten & Medien herunterladen")
		} else {
			btnStart.SetText("1. CAR-Datei verarbeiten")
		}
	})
	chkDownloadMedia.SetChecked(true) // Standardmäßig aktiviert

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)

	statusLabel := widget.NewLabel("Bereit.")
	statusLabel.Wrapping = fyne.TextWrapWord



	// --- BEREINIGEN-BEREICH ---
	chkPosts := widget.NewCheck("Posts", nil)
	chkPosts.SetChecked(true)
	btnCleanPostsTXT := widget.NewButton("Als TXT...", nil)
	btnCleanPostsHTML := widget.NewButton("Als HTML (klickbar)...", nil)

	chkFollows := widget.NewCheck("Follows", nil)
	btnCleanFollows := widget.NewButton("Bereinigen...", nil)

	chkBlocks := widget.NewCheck("Blocks", nil)
	btnCleanBlocks := widget.NewButton("Bereinigen...", nil)

	chkLikes := widget.NewCheck("Likes", nil)
	btnCleanLikes := widget.NewButton("Bereinigen...", nil)

	// Hilfsfunktion zum Aktivieren / Deaktivieren aller Bereinigen-Buttons
	setCleanButtonsState := func(enabled bool) {
		if enabled {
			if chkPosts.Checked {
				btnCleanPostsTXT.Enable()
				btnCleanPostsHTML.Enable()
			}
			if chkFollows.Checked {
				btnCleanFollows.Enable()
			}
			if chkBlocks.Checked {
				btnCleanBlocks.Enable()
			}
			if chkLikes.Checked {
				btnCleanLikes.Enable()
			}
		} else {
			btnCleanPostsTXT.Disable()
			btnCleanPostsHTML.Disable()
			btnCleanFollows.Disable()
			btnCleanBlocks.Disable()
			btnCleanLikes.Disable()
		}
	}

	// Dynamic Checkbox-Steuerung
	chkPosts.OnChanged = func(b bool) {
		if b {
			btnCleanPostsTXT.Enable()
			btnCleanPostsHTML.Enable()
		} else {
			btnCleanPostsTXT.Disable()
			btnCleanPostsHTML.Disable()
		}
	}
	chkFollows.OnChanged = func(b bool) { if b { btnCleanFollows.Enable() } else { btnCleanFollows.Disable() } }
	chkBlocks.OnChanged = func(b bool) { if b { btnCleanBlocks.Enable() } else { btnCleanBlocks.Disable() } }
	chkLikes.OnChanged = func(b bool) { if b { btnCleanLikes.Enable() } else { btnCleanLikes.Disable() } }

	// Event-Handler für TXT-Exports
	handleExport := func(defaultName string, exportFunc func(string, string) error) {
		outDir := outPathEntry.Text
		go func() {
			statusLabel.SetText("Öffne Speichern-Dialog...")
			filter := "Textdateien (*.txt)|*.txt|Alle Dateien (*.*)|*.*"

			targetPath, ok := openNativeSaveDialog(defaultName, filter)
			if !ok || targetPath == "" {
				statusLabel.SetText("Speichern abgebrochen.")
				return
			}

			statusLabel.SetText("Erstelle Textdatei...")
			err := exportFunc(outDir, targetPath)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Fehler: %v", err))
			} else {
				statusLabel.SetText(fmt.Sprintf("Erfolgreich gespeichert unter: %s", filepath.Base(targetPath)))
			}
		}()
	}

	btnCleanPostsTXT.OnTapped = func() { handleExport("posts_lesbar.txt", exportPostsAsCleanText) }

	btnCleanPostsHTML.OnTapped = func() {
		outDir := outPathEntry.Text
		isMediaChecked := chkDownloadMedia.Checked

		go func() {
			statusLabel.SetText("Öffne Speichern-Dialog für HTML...")
			filter := "HTML-Dateien (*.html)|*.html|Alle Dateien (*.*)|*.*"

			targetPath, ok := openNativeSaveDialog("posts_lesbar.html", filter)
			if !ok || targetPath == "" {
				statusLabel.SetText("Speichern abgebrochen.")
				return
			}

			statusLabel.SetText("Erstelle HTML-Dokument...")
			err := exportPostsAsHTML(outDir, targetPath, isMediaChecked)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Fehler: %v", err))
			} else {
				statusLabel.SetText(fmt.Sprintf("HTML erfolgreich gespeichert unter: %s", filepath.Base(targetPath)))
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

		if carPath == "" {
			dialog.ShowError(fmt.Errorf("Bitte wähle eine .car-Datei aus!"), myWindow)
			return
		}

		if chkDownloadMedia.Checked && handleInput == "" {
			dialog.ShowError(fmt.Errorf("Für den Bilder-Download wird ein Handle oder eine DID benötigt!"), myWindow)
			return
		}

		btnStart.Disable()
		setCleanButtonsState(false)
		progressBar.SetValue(0)

		go func() {
			statusLabel.SetText("Starte CAR-Parsing...")
			err := parseCarFile(carPath, outDir, func(msg string) {
				statusLabel.SetText(msg)
			})

			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Fehler beim Parsing: %v", err))
				btnStart.Enable()
				setCleanButtonsState(true)
				return
			}

			if chkDownloadMedia.Checked {
				statusLabel.SetText("Löse Handle/DID auf...")
				userDid, err := resolveHandleToDID(handleInput, pds)
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("Fehler bei Handle-Auflösung: %v", err))
					btnStart.Enable()
					setCleanButtonsState(true)
					return
				}

				statusLabel.SetText(fmt.Sprintf("DID gefunden (%s). Starte Medien-Download...", userDid))

				err = downloadMedia(outDir, userDid, pds, func(current, total int, msg string) {
					if total > 0 {
						progressBar.SetValue(float64(current) / float64(total))
					}
					statusLabel.SetText(msg)
				})

				if err != nil {
					statusLabel.SetText(fmt.Sprintf("Download-Fehler: %v", err))
				} else {
					statusLabel.SetText("Fertig! Alle JSONs und Medien wurden erfolgreich gespeichert.")
				}
			} else {
				progressBar.SetValue(1.0)
				statusLabel.SetText("Fertig! JSON-Dateien wurden erfolgreich exportiert (Bilder übersprungen).")
			}

			btnStart.Enable()
			setCleanButtonsState(true)
		}()
	}

	carRow := container.NewBorder(nil, nil, nil, btnBrowseCar, carPathEntry)
	outRow := container.NewBorder(nil, nil, nil, btnBrowseOut, outPathEntry)

	form := widget.NewForm(
		widget.NewFormItem("CAR-Datei:", carRow),
		widget.NewFormItem("User Handle / DID:", handleEntry),
		widget.NewFormItem("PDS Server:", pdsEntry),
		widget.NewFormItem("Zielordner:", outRow),
		widget.NewFormItem("Optionen:", chkDownloadMedia),
	)

	postsButtons := container.NewGridWithColumns(2, btnCleanPostsTXT, btnCleanPostsHTML)

	cleanGrid := container.NewGridWithColumns(2,
		chkPosts, postsButtons,
		chkFollows, btnCleanFollows,
		chkBlocks, btnCleanBlocks,
		chkLikes, btnCleanLikes,
	)

	// Statusleiste: Links Status, Rechts Credits
	bottomBar := container.NewBorder(nil, nil, nil, creditsLabel, statusLabel)

	// Gesamter Inhalt in einer vertikalen Box
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		btnStart,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("2. Daten bereinigen & als Text/HTML speichern:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		cleanGrid,
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