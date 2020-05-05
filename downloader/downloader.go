package downloader

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareaderdl/fetch"
	"github.com/francoiscolombo/gomangareaderdl/settings"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type DownloadItem struct {
	Manga string
	Chapter int
	Provider string
	OutputPath string
}

/*
cleanup is called when we close the GUI
 */
func cleanup() {
	fmt.Println("--- GUI exited")
}


/*
Download allow to display a GUI to download a manga instead of using the command line.
 */
func Download(cfg *settings.Settings) error {
	downloader := app.New()

	w := downloader.NewWindow("Download mangas")

	// status bar items
	status := widget.NewLabel("-")

	// since we can't receive the parameters from the command line, we have to ask the user to enter them
	mangaTitleLabel := widget.NewLabel("Which manga to download? ")
	mangaTitle := widget.NewEntry()
	mangaTitle.SetPlaceHolder("Enter manga title...")
	var titles []string
	for _, manga := range cfg.History.Titles {
		titles = append(titles, manga.Title)
	}
	newMangaTitle := widget.NewSelect(titles, func(value string) {
		mangaTitle.SetText(value)
	})
	mangaTitleSelector := widget.NewHBox(mangaTitle, newMangaTitle)

	mangaProviderLabel := widget.NewLabel("Select a provider: ")
	mangaProvider := widget.NewSelect([]string{"mangapanda.com","mangareader.net","mangalife.us"}, func(value string) {
		//statusProvider.SetText(value)
	})
	mangaProvider.Selected = cfg.Config.Provider

	mangaChapterLabel := widget.NewLabel("Starting from which chapter? ")
	mangaChapter := widget.NewEntry()
	mangaChapter.SetText("1")
	mangaChapterIncreaseButton := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		v, _ := strconv.Atoi(mangaChapter.Text)
		v = v+1
		mangaChapter.SetText(strconv.Itoa(v))
	})
	mangaChapterDecreaseButton := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		v, _ := strconv.Atoi(mangaChapter.Text)
		v = v-1
		if v<0 {
			v=0
		}
		mangaChapter.SetText(strconv.Itoa(v))
	})
	mangaChapterSelector := widget.NewHBox(mangaChapter,mangaChapterIncreaseButton,mangaChapterDecreaseButton)

	mangaPathLabel := widget.NewLabel("Where do you want to store your mangas? ")
	mangaPath := widget.NewEntry()
	mangaPath.SetPlaceHolder("Enter where to store the mangas downloaded...")
	mangaPath.SetText(cfg.Config.OutputPath)

	mangaContinueCheckLabel := widget.NewLabel("no, single chapter download")
	mangaContinueCheck := widget.NewCheck("Download continuously all chapters?", func(isChecked bool) {
		if isChecked {
			mangaContinueCheckLabel.SetText("yes, do continuous downloads")
		} else {
			mangaContinueCheckLabel.SetText("no, single chapter download")
		}
	})
	formParams := fyne.NewContainerWithLayout(layout.NewFormLayout(),
		mangaTitleLabel, mangaTitleSelector,
		mangaProviderLabel, mangaProvider,
		mangaChapterLabel, mangaChapterSelector,
		mangaContinueCheck, mangaContinueCheckLabel,
		mangaPathLabel, mangaPath)

	// form to display what we are currently downloading
	mangaCurrentOutputName := widget.NewLabel("")
	mangaCurrentProgress := widget.NewProgressBar()
	mangaCurrentProgress.Resize(fyne.NewSize(600,20))
	mangaCurrentlyDownloaded := widget.NewHBox(widget.NewLabel("Currently downloading: "),
		mangaCurrentOutputName, mangaCurrentProgress)

	params := widget.NewVBox(formParams, widget.NewButtonWithIcon("Download", theme.DocumentSaveIcon(), func() {
		downloadingChapter, _ := strconv.Atoi(mangaChapter.Text)
		for {
			status.SetText(fmt.Sprintf("Now downloading chapter %d for %s from %s... Please wait.",
				downloadingChapter, mangaTitle.Text, mangaProvider.Selected))
			mangaCurrentOutputName.SetText(fmt.Sprintf("%s-%03d.cbz",mangaTitle.Text,downloadingChapter))
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				downloadPath := fmt.Sprintf("%s/%s/%03d", mangaPath.Text, mangaTitle.Text, downloadingChapter)
				cbzPath := fmt.Sprintf("%s/%s", mangaPath.Text, mangaTitle.Text)
				// check if the download dir exist, and if not create it
				if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
					_ = os.MkdirAll(downloadPath, os.ModePerm)
				}
				// erase the content before (in case the directory was already existing)
				dir, _ := ioutil.ReadDir(downloadPath)
				for _, d := range dir {
					_ = os.RemoveAll(path.Join([]string{downloadPath, d.Name()}...))
				}
				count, imgURL := fetch.SearchPages(mangaProvider.Selected, mangaTitle.Text, downloadingChapter)
				currentPage := 0
				for p, img := range imgURL {
					fetch.DownloadImage(downloadPath, p, fetch.SearchImage(mangaProvider.Selected, mangaTitle.Text, img))
					currentPage = currentPage + 1
					mangaCurrentProgress.SetValue(float64(currentPage) / float64(count))
					// being nice and let the time to the GUI to refresh
					time.Sleep(time.Millisecond * 250)
				}
				mangaCurrentProgress.SetValue(1.0)
				status.SetText(fmt.Sprintf("Chapter %d of %s downloaded. Now, creating the CBZ...",
					downloadingChapter, mangaTitle.Text))
				fetch.CreateCBZ(cbzPath, downloadPath, mangaTitle.Text, downloadingChapter)
				status.SetText("CBZ created.")
				wg.Done()
			}()
			wg.Wait()
			if mangaContinueCheck.Checked {
				downloadingChapter = downloadingChapter + 1
				if fetch.NextChapter(mangaProvider.Selected,mangaTitle.Text,downloadingChapter) == false {
					mangaContinueCheck.SetChecked(false)
					break
				}
				mangaChapter.SetText(strconv.Itoa(downloadingChapter))
			} else {
				break
			}
		}
	}), mangaCurrentlyDownloaded)

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, status, nil, nil), status, params)

	w.SetContent(content)
	w.Show()

	downloader.Run()
	cleanup()

	return nil
}
