package fetch

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/francoiscolombo/gomangareaderdl/createcbz"
	"github.com/schollz/progressbar"
	"gopkg.in/gookit/color.v1"
)

func createCBZ(outputPath, pagesPath, title string, chapter int) {
	cstep := color.FgYellow.Render
	cprm := color.FgLightMagenta.Render
	cdone := color.FgLightGreen.Render
	// List of Files to Zip
	fmt.Printf("\n%s %s ... ", cstep("create"), cprm(fmt.Sprintf("%s-%03d.cbz", title, chapter)))
	var files []string
	outputCBZ := fmt.Sprintf("%s/%s-%03d.cbz", outputPath, title, chapter)
	err := filepath.Walk(pagesPath, func(path string, info os.FileInfo, err error) error {
		src, err := os.Stat(path)
		if err != nil {
			// still does not exists? then something wrong, exit in panic mode.
			panic(err)
		}
		if !src.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	if err := createcbz.ZipFiles(outputCBZ, files); err != nil {
		panic(err)
	}
	for _, file := range files {
		os.Remove(file)
	}
	os.Remove(pagesPath)
	fmt.Printf("%s\n", cdone("done"))
}

func downloadImage(path string, page int, url string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to extract images from %s: %d %s", url, res.StatusCode, res.Status)
	}
	//open a file for writing
	file, err := os.Create(fmt.Sprintf("%s/page_%03d.jpg", path, page))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func searchImage(provider, title, url string) (imageURL string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to extract images from %s: %d %s", url, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	r, _ := regexp.Compile(fmt.Sprintf("^http.*://.+\\.%s/%s/\\d+/.+\\..+$", provider, title))
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		v, _ := s.Attr("src")
		if r.MatchString(v) {
			imageURL = v
			return
		}
	})
	return
}

func searchPages(provider, title string, chapter int) (count int, imagesURL []string) {
	url := fmt.Sprintf("https://www.%s/%s/%d/%d", provider, title, chapter, 1)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to extract images from %s: %d %s", url, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	// count how many images
	count = 0
	doc.Find("option").Each(func(i int, s *goquery.Selection) {
		v, _ := s.Attr("value")
		imagesURL = append(imagesURL, fmt.Sprintf("https://www.%s%s", provider, v))
		count = count + 1
	})
	return
}

func downloadChapter(path, provider, title string, chapter int, displayProgressBar bool) {
	cstep := color.FgYellow.Render
	cprm := color.FgLightMagenta.Render
	cdone := color.FgLightGreen.Render
	if displayProgressBar {
		fmt.Printf("%s to download ... ", cstep("search pages"))
	}
	count, imgURL := searchPages(provider, title, chapter)
	if displayProgressBar {
		fmt.Printf("%s (found %s pages for %s chapter %s)\n", cdone("done"), cprm(fmt.Sprintf("%d", count)), cprm(title), cprm(fmt.Sprintf("%d", chapter)))
		// and then search for images to download
		fmt.Printf("%s pages ...\n", cstep("download"))
		bar := progressbar.NewOptions(count)
		bar.RenderBlank()
		var wg sync.WaitGroup
		wg.Add(len(imgURL))
		for p, img := range imgURL {
			go func(page int, urlImg string) {
				downloadImage(path, page, searchImage(provider, title, urlImg))
				bar.Add(1)
				wg.Done()
			}(p, img)
		}
		wg.Wait()
	} else {
		for p, img := range imgURL {
			downloadImage(path, p, searchImage(provider, title, img))
		}

	}
	return
}

/*
Manga download a manga
*/
func Manga(provider, title string, chapter int, outputPath string, displayProgressBar bool) (nextChapter int) {
	// check if download path exists
	downloadPath := fmt.Sprintf("%s/%s/%03d", outputPath, title, chapter)
	cbzPath := fmt.Sprintf("%s/%s", outputPath, title)
	// check if the download dir exist, and if not create it
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		os.MkdirAll(downloadPath, os.ModePerm)
	}
	src, err := os.Stat(downloadPath)
	if err != nil {
		// still does not exists? then something wrong, exit in panic mode.
		panic(err)
	}
	// check if the source is indeed a directory or not
	if !src.IsDir() {
		color.Error.Prompt("source path is not a directory")
		os.Exit(1)
	} else {
		// it's a directory, then erase the content before
		dir, _ := ioutil.ReadDir(downloadPath)
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{downloadPath, d.Name()}...))
		}
	}
	downloadChapter(downloadPath, provider, title, chapter, displayProgressBar)
	createCBZ(cbzPath, downloadPath, title, chapter)
	nextChapter = chapter + 1
	return
}

/*
NextChapter check if a new chapter exists, return true if exists and false otherwise
*/
func NextChapter(provider, title string, chapter int) bool {
	count, _ := searchPages(provider, title, chapter)
	return (count > 0)
}
