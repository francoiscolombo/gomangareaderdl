package commands

import (
	"fmt"
	"os"

	"github.com/francoiscolombo/gomangareaderdl/fetch"
	"github.com/francoiscolombo/gomangareaderdl/settings"
)

/*
ProcessConfigCommand process the config command, and update the default configuration regarding the parameters passed
*/
func ProcessConfigCommand(cfg *settings.Settings, defaultOutputPath string, defaultProvider string) {
	if defaultOutputPath == "???" {
		defaultOutputPath = cfg.Config.OutputPath
	}
	if defaultProvider == "???" {
		defaultProvider = cfg.Config.Provider
	}
	fmt.Println("- <Config> command selected, with the following parameters:")
	fmt.Printf("  > Default output path to set : '%s'\n", defaultOutputPath)
	fmt.Printf("  > Default provider is set to <%s>", defaultProvider)
	if (defaultOutputPath != cfg.Config.OutputPath) || (defaultProvider != cfg.Config.Provider) {
		(*cfg).Config.OutputPath = defaultOutputPath
		(*cfg).Config.Provider = defaultProvider
		settings.WriteSettings((*cfg))
	}
}

/*
ProcessFetchCommand allows to download a manga, from the first given chapter to the last available one.
*/
func ProcessFetchCommand(cfg *settings.Settings, manga string, chapter int, provider string, path string, force bool, silent bool) {
	if manga == "???" {
		fmt.Println("parameter --manga is mandatory...")
		os.Exit(1)
	}
	if path == "???" {
		path = cfg.Config.OutputPath
	}
	if provider == "???" {
		provider = cfg.Config.Provider
	}
	if chapter < 0 {
		chapter = settings.SearchLastChapter((*cfg), manga)
	}
	fmt.Println("- <Fetch> command selected, with the following parameters:")
	fmt.Printf("  > Manga title to fetch : '%s'\n", manga)
	fmt.Printf("  > Download from provider <%s>\n", provider)
	fmt.Printf("  > Start to fetch from chapter %d\n", chapter)
	fmt.Printf("  > Download to output path '%s'\n", path)
	if force {
		fmt.Printf("  > We are restarting the download from chapter %d\n", chapter)
	} else {
		lastChapter := settings.SearchLastChapter((*cfg), manga)
		if lastChapter > chapter {
			chapter = lastChapter
		}
		fmt.Printf("  > We are now searching for new chapter %d\n", chapter)
	}
	if silent {
		fmt.Printf("  > Download of %s will be done silently (no progress bar)\n", manga)
	}
	if fetch.NextChapter(provider, manga, chapter) == false {
		fmt.Printf("chapter %d for %s is not yet available to download, sorry.", chapter, manga)
	} else {
		for {
			if fetch.NextChapter(provider, manga, chapter) == true {
				chapter = fetch.Manga(provider, manga, chapter, path, !silent)
			} else {
				break
			}
		}
		*cfg = settings.UpdateHistory(*cfg, manga, chapter, provider)
	}
}

/*
ProcessListCommand process the list command, highlight the mangas that have new chapters for all the suscribed
mangas available in the history
*/
func ProcessListCommand(cfg *settings.Settings) {
	fmt.Println("- <List> command selected")
	settings.DisplayHistory(cfg)
}

/*
ProcessUpdateCommand allows to update the history for a downloaded manga. you can override
the provider, or the next chapter to download.
*/
func ProcessUpdateCommand(cfg *settings.Settings, manga, provider string, nextChapter int) {
	if manga == "???" {
		fmt.Println("parameter --manga is mandatory...")
		os.Exit(1)
	}
	fmt.Println("- <Update> command selected, with the following parameters:")
	fmt.Printf("  > Filter on Manga title : '%s'\n", manga)
	if provider == "???" {
		fmt.Printf("  > Set provider to : '%s'\n", provider)
	}
	if nextChapter > 0 {
		fmt.Printf("  > Set next chapter to download to %d\n", nextChapter)
	}
	*cfg = settings.UpdateHistory(*cfg, manga, nextChapter, provider)
}