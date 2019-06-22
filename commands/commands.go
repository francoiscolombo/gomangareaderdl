package commands

import (
	"fmt"
	"os"

	"github.com/francoiscolombo/gomangareaderdl/fetch"
	"github.com/francoiscolombo/gomangareaderdl/settings"
	"github.com/francoiscolombo/gomangareaderdl/viewer"

	"gopkg.in/gookit/color.v1"
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
	ccmd := color.FgLightBlue.Render
	cprm := color.FgLightCyan.Render
	fmt.Printf("- %s command selected, with the following parameters:\n", ccmd("Config"))
	fmt.Printf("  > Default output path to set : '%s'\n", cprm(defaultOutputPath))
	fmt.Printf("  > Default provider is set to <%s>", cprm(defaultProvider))
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
		color.Error.Prompt("parameter --manga is mandatory...")
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
	ccmd := color.FgLightBlue.Render
	cprm := color.FgLightCyan.Render
	cwarn := color.FgLightYellow.Render
	fmt.Printf("- %s command selected, with the following parameters:\n", ccmd("Fetch"))
	fmt.Printf("  > Manga title to fetch : '%s'\n", cprm(manga))
	fmt.Printf("  > Download from provider <%s>\n", cprm(provider))
	fmt.Printf("  > Start to fetch from chapter %s\n", cprm(fmt.Sprintf("%d", chapter)))
	fmt.Printf("  > Download to output path '%s'\n", cprm(path))
	if force {
		fmt.Printf("  > We are restarting the download from chapter %s\n", cprm(fmt.Sprintf("%d", chapter)))
	} else {
		lastChapter := settings.SearchLastChapter((*cfg), manga)
		if lastChapter > chapter {
			chapter = lastChapter
		}
		fmt.Printf("  > We are now searching for new chapter %s\n", cprm(fmt.Sprintf("%d", chapter)))
	}
	if silent {
		fmt.Printf("  > Download of %s will be done silently (no progress bar)\n", cprm(manga))
	}
	if fetch.NextChapter(provider, manga, chapter) == false {
		color.Warn.Tips(fmt.Sprintf("chapter %s for %s is not yet available to download, sorry.", cwarn(fmt.Sprintf("%d", chapter)), cwarn(manga)))
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
	ccmd := color.FgLightBlue.Render
	fmt.Printf("- %s command selected\n", ccmd("List"))
	settings.DisplayHistory(cfg)
}

/*
ProcessUpdateCommand allows to update the history for a downloaded manga. you can override
the provider, or the next chapter to download.
*/
func ProcessUpdateCommand(cfg *settings.Settings, manga, provider string, nextChapter int) {
	if manga == "???" {
		color.Error.Prompt("parameter --manga is mandatory...")
		os.Exit(1)
	}
	ccmd := color.FgLightBlue.Render
	cprm := color.FgLightCyan.Render
	fmt.Printf("- %s command selected, with the following parameters:\n", ccmd("Update"))
	fmt.Printf("  > Filter on Manga title : '%s'\n", cprm(manga))
	if provider == "???" {
		fmt.Printf("  > Set provider to : '%s'\n", cprm(provider))
	}
	if nextChapter > 0 {
		fmt.Printf("  > Set next chapter to download to %s\n", cprm(fmt.Sprintf("%d", nextChapter)))
	}
	*cfg = settings.UpdateHistory(*cfg, manga, nextChapter, provider)
}

/*
ProcessViewCommand allows to view a previously downloaded manga given chapter
*/
func ProcessViewCommand(cfg *settings.Settings, manga string, chapter int, path string) {
	if manga == "???" {
		color.Error.Prompt("parameter --manga is mandatory...")
		os.Exit(1)
	}
	if chapter < 0 {
		chapter = settings.SearchLastChapter((*cfg), manga)
	}
	if path == "???" {
		path = cfg.Config.OutputPath
	}
	ccmd := color.FgLightBlue.Render
	cprm := color.FgLightCyan.Render
	cwarn := color.FgLightYellow.Render
	fmt.Printf("- %s command selected, with the following parameters:\n", ccmd("View"))
	fmt.Printf("  > Manga title to view : '%s'\n", cprm(manga))
	fmt.Printf("  > Read chapter %s\n", cprm(fmt.Sprintf("%d", chapter)))
	fmt.Printf("  > From path '%s'\n", cprm(path))
	err := viewer.Read(manga, chapter, path)
	if err != nil {
		color.Error.Tips(fmt.Sprintf("Error when trying to open manga %s chapter %s for reading: %s", cwarn(manga), cwarn(fmt.Sprintf("%d", chapter)), err))
	}
}
