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
	text := textsFor(detectUILanguage())
	myApp := app.New()
	myApp.SetIcon(fyne.NewStaticResource("icon.png", appIcon))
	myWindow := myApp.NewWindow(text.windowTitle)
	myWindow.Resize(fyne.NewSize(680, 620))
	creditsLabel := widget.NewLabelWithStyle("v1.1.0 by Zamara74", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	carPathEntry := widget.NewEntry()
	carPathEntry.SetPlaceHolder(text.carPathPlaceholder)

	btnBrowseCar := widget.NewButton(text.browseFiles, func() {
		go func() {
			if selectedFile, ok := openNativeOpenFileDialog(text.carFileFilter, text.selectCarTitle); ok {
				fyne.Do(func() {
					carPathEntry.SetText(selectedFile)
				})
			}
		}()
	})

	handleEntry := widget.NewEntry()
	handleEntry.SetPlaceHolder(text.handlePlaceholder)

	chatPasswordEntry := widget.NewPasswordEntry()
	chatPasswordEntry.SetPlaceHolder(text.appPasswordPlaceholder)

	pdsEntry := widget.NewEntry()
	pdsEntry.SetPlaceHolder(text.pdsPlaceholder)
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

	btnBrowseOut := widget.NewButton(text.browseFolder, func() {
		go func() {
			if selectedFolder, ok := openNativeFolderDialog(text.selectFolderDescription); ok {
				fyne.Do(func() {
					outPathEntry.SetText(selectedFolder)
				})
			}
		}()
	})

	// 1. Button zuerst mit Standard-Text und HighImportance erstellen
	btnStart := widget.NewButton(text.startWithMedia, func() {
		// Hier kommt deine bisherige Start-Funktion/Logik rein
	})
	btnStart.Importance = widget.HighImportance

	// 2. Eine EINZIGE Checkbox erstellen, die Text-Wechsel UND Haken steuert
	chkDownloadMedia := widget.NewCheck(text.downloadImages, func(checked bool) {
		if checked {
			btnStart.SetText(text.startWithMedia)
		} else {
			btnStart.SetText(text.startWithoutMedia)
		}
	})
	chkDownloadMedia.SetChecked(true) // Standardmäßig aktiviert

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)

	statusLabel := widget.NewLabel(text.statusReady)
	statusLabel.Wrapping = fyne.TextWrapWord
	setStatus := func(message string) {
		fyne.Do(func() {
			statusLabel.SetText(message)
		})
	}

	// --- EXPORT-BEREICH ---
	btnCleanPostsTXT := widget.NewButton(text.exportAsText, nil)
	btnCleanPostsHTML := widget.NewButton(text.exportAsHTMLFile, nil)

	btnCleanFollows := widget.NewButton(text.exportAsText, nil)
	btnCleanBlocks := widget.NewButton(text.exportAsText, nil)
	btnCleanLikes := widget.NewButton(text.exportAsText, nil)

	// Button für den Online-Chat-Export
	backupRunning := false
	var btnCleanChats *widget.Button
	btnCleanChats = widget.NewButton(text.exportAsHTML, func() {
		if handleEntry.Text == "" {
			dialog.ShowInformation(text.errorTitle, text.missingHandle, myWindow)
			return
		}
		if chatPasswordEntry.Text == "" {
			dialog.ShowInformation(text.errorTitle, text.missingPassword, myWindow)
			return
		}
		if pdsEntry.Text == "" {
			dialog.ShowInformation(text.pdsRequiredTitle, text.pdsRequiredMessage, myWindow)
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
		statusLabel.SetText(text.loginDownloadingChats)

		go func() {
			conversations, myDID, err := FetchChatConversations(handle, appPassword, pds)
			appPassword = ""
			if err != nil {
				setStatus(fmt.Sprintf(text.chatDownloadError, err))
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
				setStatus(fmt.Sprintf(text.chatHTMLSaveError, err))
				fyne.Do(func() {
					if !backupRunning {
						btnCleanChats.Enable()
					}
				})
				return
			}

			setStatus(fmt.Sprintf(text.chatsSaved, targetPath))
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
			setStatus(text.openSaveDialog)
			filter := text.textFileFilter

			targetPath, ok := openNativeSaveDialog(defaultName, filter, text.allFilesFilter, text.saveFileTitle)
			if !ok || targetPath == "" {
				setStatus(text.saveCancelled)
				return
			}

			setStatus(text.creatingTextFile)
			err := exportFunc(outDir, targetPath)
			if err != nil {
				setStatus(fmt.Sprintf("%s: %v", text.errorTitle, err))
			} else {
				setStatus(fmt.Sprintf(text.savedSuccessfully, filepath.Base(targetPath)))
			}
		}()
	}

	btnCleanPostsTXT.OnTapped = func() { handleExport("posts_lesbar.txt", exportPostsAsCleanText) }

	btnCleanPostsHTML.OnTapped = func() {
		outDir := outPathEntry.Text
		isMediaChecked := chkDownloadMedia.Checked

		go func() {
			setStatus(text.openHTMLSaveDialog)
			filter := text.htmlFileFilter

			targetPath, ok := openNativeSaveDialog("posts_lesbar.html", filter, text.allFilesFilter, text.saveFileTitle)
			if !ok || targetPath == "" {
				setStatus(text.saveCancelled)
				return
			}

			setStatus(text.creatingHTML)
			err := exportPostsAsHTML(outDir, targetPath, isMediaChecked)
			if err != nil {
				setStatus(fmt.Sprintf("%s: %v", text.errorTitle, err))
			} else {
				setStatus(fmt.Sprintf(text.htmlSavedSuccessfully, filepath.Base(targetPath)))
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
			dialog.ShowError(fmt.Errorf("%s", text.missingCAR), myWindow)
			return
		}

		if downloadMediaChecked && handleInput == "" {
			dialog.ShowError(fmt.Errorf("%s", text.handleRequiredForMedia), myWindow)
			return
		}
		if downloadMediaChecked && pdsEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("%s", text.pdsNotRecognized), myWindow)
			return
		}

		btnStart.Disable()
		backupRunning = true
		setExportButtonsState(false)
		progressBar.SetValue(0)

		go func() {
			setStatus(text.startingCARParsing)
			err := parseCarFile(carPath, outDir, text, func(msg string) {
				setStatus(msg)
			})

			if err != nil {
				setStatus(fmt.Sprintf(text.parsingError, err))
				fyne.Do(func() {
					btnStart.Enable()
				})
				setExportButtonsState(true)
				return
			}

			if downloadMediaChecked {
				setStatus(text.resolvingHandle)
				userDid, err := resolveHandleToDID(handleInput, pds)
				if err != nil {
					setStatus(fmt.Sprintf(text.handleResolutionError, err))
					fyne.Do(func() {
						btnStart.Enable()
					})
					setExportButtonsState(true)
					return
				}

				setStatus(fmt.Sprintf(text.didFoundStartingDownload, userDid))

				err = downloadMedia(outDir, userDid, pds, text, func(current, total int, msg string) {
					setStatus(msg)
					if total > 0 {
						value := float64(current) / float64(total)
						fyne.Do(func() {
							progressBar.SetValue(value)
						})
					}
				})

				if err != nil {
					setStatus(fmt.Sprintf(text.downloadError, err))
				} else {
					setStatus(text.backupComplete)
				}
			} else {
				fyne.Do(func() {
					progressBar.SetValue(1.0)
				})
				setStatus(text.exportComplete)
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
		widget.NewFormItem(text.carFileLabel, carRow),
		widget.NewFormItem(text.handleDIDLabel, handleEntry),
		widget.NewFormItem(text.appPasswordLabel, chatPasswordEntry),
		widget.NewFormItem(text.pdsLabel, pdsEntry),
		widget.NewFormItem(text.outputFolderLabel, outRow),
		widget.NewFormItem(text.optionsLabel, chkDownloadMedia),
	)

	postsButtons := container.NewHBox(btnCleanPostsTXT, btnCleanPostsHTML)

	leftExportColumn := container.New(layout.NewFormLayout(),
		widget.NewLabel(text.postsLabel), postsButtons,
		widget.NewLabel(text.followsLabel), container.NewHBox(btnCleanFollows),
		widget.NewLabel(text.blocksLabel), container.NewHBox(btnCleanBlocks),
	)
	rightExportColumn := container.New(layout.NewFormLayout(),
		widget.NewLabel(text.likesLabel), container.NewHBox(btnCleanLikes),
		widget.NewLabel(text.chatsLabel), container.NewHBox(btnCleanChats),
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
		widget.NewLabelWithStyle(text.cleanDataLabel, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
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
