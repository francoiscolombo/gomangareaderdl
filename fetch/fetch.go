package fetch

import (
	"fmt"
	"github.com/francoiscolombo/gomangareaderdl/createcbz"
	"github.com/schollz/progressbar/v2"
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
)

/*
CreateCBZ create a readable comics archive from the pages downloaded and clean the temporary directory
 */
func CreateCBZ(outputPath, pagesPath, title string, chapter int) {
	// List of Files to Zip
	fmt.Printf("\ncreate %s ... ", fmt.Sprintf("%s-%03d.cbz", title, chapter))
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
	fmt.Println("done")
}

/*
DownloadImage simply download an image and store it in the proper directory
 */
func DownloadImage(path string, page int, url string) {
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

/*
SearchImage search in HTML page all the link that respect the pattern expected for downloading a comic page
 */
func SearchImage(provider, title, url string) (imageURL string) {
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

/*
SearchPages will send the number of page to download, and a list of url for every image to download
 */
func SearchPages(provider, title string, chapter int) (count int, imagesURL []string) {
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
	if displayProgressBar {
		fmt.Printf("search pages to download ... ")
	}
	count, imgURL := SearchPages(provider, title, chapter)
	if displayProgressBar {
		fmt.Printf("done (found %d pages for %s chapter %d)\n", count, title, chapter)
		// and then search for images to download
		fmt.Println("download pages ...")
		bar := progressbar.NewOptions(count)
		bar.RenderBlank()
		var wg sync.WaitGroup
		wg.Add(len(imgURL))
		for p, img := range imgURL {
			go func(page int, urlImg string) {
				DownloadImage(path, page, SearchImage(provider, title, urlImg))
				bar.Add(1)
				wg.Done()
			}(p, img)
		}
		wg.Wait()
	} else {
		for p, img := range imgURL {
			DownloadImage(path, p, SearchImage(provider, title, img))
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
		fmt.Println("source path is not a directory")
		os.Exit(1)
	} else {
		// it's a directory, then erase the content before
		dir, _ := ioutil.ReadDir(downloadPath)
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{downloadPath, d.Name()}...))
		}
	}
	downloadChapter(downloadPath, provider, title, chapter, displayProgressBar)
	CreateCBZ(cbzPath, downloadPath, title, chapter)
	nextChapter = chapter + 1
	return
}

/*
NextChapter check if a new chapter exists, return true if exists and false otherwise
*/
func NextChapter(provider, title string, chapter int) bool {
	count, _ := SearchPages(provider, title, chapter)
	return (count > 0)
}
