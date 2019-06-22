package viewer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func unzip(src string, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, filepath.Base(f.Name))
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, fpath)
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

/*
Read open the GUI allowing to watch a selected chapter of a manga
*/
func Read(manga string, chapter int, path string) error {

	fromCbz := fmt.Sprintf("%s%s%s%s%s-%03d.cbz", path, string(os.PathSeparator), manga, string(os.PathSeparator), manga, chapter)
	tmpDir := fmt.Sprintf("%s%stmpviewer%s%s-%03d", path, string(os.PathSeparator), string(os.PathSeparator), manga, chapter)
	pages, err := unzip(fromCbz, tmpDir)
	if err != nil {
		return err
	}
	pageNumber := 1
	nbPages := len(pages)

	app := app.New()

	w := app.NewWindow(fmt.Sprintf("View Chapter %d of %s", chapter, manga))

	displayPage := widget.NewLabel(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
	pageProgress := widget.NewProgressBar()
	pageProgress.SetValue(float64(pageNumber) / float64(nbPages))

	pageView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	pageView.File = pages[pageNumber-1]
	canvas.Refresh(pageView)

	navBar := fyne.NewContainerWithLayout(
		layout.NewGridLayout(2),
		widget.NewHBox(
			widget.NewButton("Quit", func() {
				app.Quit()
			}),
			layout.NewSpacer(),
			widget.NewButton("<<< [Prev]", func() {
				pageNumber--
				if pageNumber < 1 {
					pageNumber = 1
				}
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
			}),
			widget.NewButton("[Next] >>>", func() {
				pageNumber++
				if pageNumber > nbPages {
					pageNumber = nbPages
				}
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
			}),
			layout.NewSpacer(),
			displayPage,
		),
		pageProgress,
	)

	w.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, navBar, nil, nil), navBar, pageView))

	w.ShowAndRun()

	err = os.RemoveAll(tmpDir)

	return err
}
