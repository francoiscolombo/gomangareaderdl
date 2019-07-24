package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/francoiscolombo/gomangareaderdl/fetch"
	"github.com/olekukonko/tablewriter"
)

// Settings is the structure that allowed to store the default configuration and the download history for all the mangas we are downloading
type Settings struct {
	Config  Config  `json:"config"`
	History History `json:"history"`
}

// Config only store the default configuration, like output path, provider and if we have to use directories to store mangas
type Config struct {
	OutputPath string `json:"outputPath"`
	Provider   string `json:"provider"`
}

// History is the manga download history, so it's an array of all the mangas we are downloading
type History struct {
	Titles []Manga `json:"titles"`
}

// Manga keep the download history for every mangas that we are suscribing
type Manga struct {
	Title    string `json:"title"`
	Chapter  int    `json:"chapter"`
	Provider string `json:"provider"`
}

func getSettingsPath() string {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("Error when trying to get current user: %s\n", err)
		os.Exit(1)
	}
	return user.HomeDir + "/.gomangareader.json"
}

/*
IsSettingsExisting allows to check if the settings file already exists or no
*/
func IsSettingsExisting() bool {
	if _, err := os.Stat(getSettingsPath()); !os.IsNotExist(err) {
		return true
	}
	return false
}

/*
WriteDefaultSettings write the default settings
*/
func WriteDefaultSettings() {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("Error when trying to get current user: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Hello %s ! You don't have any settings yet. I can see that your homedir is %s, I will use it if you don't mind.\n", user.Name, user.HomeDir)
	settings := Settings{
		Config{
			OutputPath: fmt.Sprintf("%s/mangas", user.HomeDir),
			Provider:   "mangareader.net",
		},
		History{
			Titles: []Manga{},
		},
	}
	file, _ := json.MarshalIndent(settings, "", " ")
	_ = ioutil.WriteFile(getSettingsPath(), file, 0644)
}

/*
ReadSettings read the settings file
*/
func ReadSettings() (settings Settings) {

	// Open our jsonFile
	settingsPath := getSettingsPath()
	fmt.Printf("Loading settings from %s...\n", settingsPath)
	jsonFile, err := os.Open(settingsPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Error when trying to open settings file: %s\n", err)
	}

	fmt.Println("Successfully Opened settings.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'settings' which we defined above
	json.Unmarshal(byteValue, &settings)

	return
}

/*
WriteSettings write a settings file. used to change the default config or add manga to history download
*/
func WriteSettings(settings Settings) {
	file, _ := json.MarshalIndent(settings, "", " ")
	_ = ioutil.WriteFile(getSettingsPath(), file, 0644)
}

/*
SearchLastChapter send the last chapter in the history for a manga, or 1 if no history exists yet
*/
func SearchLastChapter(settings Settings, manga string) (lastChapter int) {
	lastChapter = 1
	for _, title := range settings.History.Titles {
		if title.Title == manga {
			lastChapter = title.Chapter
			break
		}
	}
	return
}

/*
DisplayHistory simply load the settings and display the titles, providers, download path and last
dowloaded chapter, and highlight mangas that have available new chapters
*/
func DisplayHistory(cfg *Settings) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Last chapter", "Provider"})
	for _, title := range (*cfg).History.Titles {
		chapter := fmt.Sprintf("%d", title.Chapter)
		mangaTitle := title.Title
		provider := title.Provider
		if fetch.NextChapter(title.Provider, title.Title, title.Chapter) == true {
			chapter = fmt.Sprintf("<%d>", title.Chapter)
			mangaTitle = fmt.Sprintf("> %s", mangaTitle)
			provider = fmt.Sprintf("[%s]", provider)
		}
		table.Append([]string{
			mangaTitle,
			chapter,
			provider,
		})
	}
	table.Render()
}

/*
UpdateHistory register the last chapter downloaded for a manga, and the last provider used
*/
func UpdateHistory(cfg Settings, manga string, chapter int, provider string) (newSettings Settings) {
	if chapter < 0 {
		chapter = 1
	}
	if provider == "???" {
		provider = cfg.Config.Provider
	}
	var titles []Manga
	for _, title := range cfg.History.Titles {
		if title.Title != manga {
			titles = append(titles, Manga{
				Title:    title.Title,
				Chapter:  title.Chapter,
				Provider: title.Provider,
			})
		}
	}
	titles = append(titles, Manga{
		Title:    manga,
		Chapter:  chapter,
		Provider: provider,
	})
	newSettings = Settings{
		Config{
			OutputPath: cfg.Config.OutputPath,
			Provider:   cfg.Config.Provider,
		},
		History{
			Titles: titles,
		},
	}
	WriteSettings(newSettings)
	fmt.Println("History updated.")
	return
}
