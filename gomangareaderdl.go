package main

import (
	"flag"
	"fmt"

	"github.com/francoiscolombo/gomangareaderdl/commands"
	"github.com/francoiscolombo/gomangareaderdl/settings"
)

const (
	versionNumber = "0.4"
	versionName   = "nebula storm"
	commandFetch  = "fetch"
	commandList   = "list"
	commandConfig = "config"
	commandUpdate = "update"
	commandView   = "view"
)

type parameters struct {
	Command  string
	Manga    string
	Chapter  int
	Provider string
	Path     string
	Silent   bool
	Force    bool
	Output   string
	Next     int
}

func usage() {
	fmt.Println(`gomangareaderdl: CLI for manga mass download

Usage
 $ gomangareaderdl -command fetch -manga <manga>

Commands list
 fetch     Fetch manga to download
 config    Set defaults
 update    Update subscribed manga
 list      List downloaded manga
 view      Launch a graphical viewer, allowing you to watch a previously downloaded manga

Options, Sub-commands
 fetch
  -manga       Set manga to download
  -chapter     Set start chapter to download
  -provider    Set download site (if not set, the default provider is used)
  -path        If used, allow to download manga to another path instead of the default one
  -force       Overwrite history
  -silent      Don't display download progress bar
 config
  -output      Set default output path
  -provider    Set default provider
 update
  -manga       Set manga to update (must have been loaded once before)
  -provider    Override download site
  -next        Set next chapter to download (rewrite history)
 view
  -manga       Set manga to view
  -chapter     Set chapter to read
  -path        If used, allow to read manga from another path instead of the default one

Example
 $ gomangareaderdl -command fetch -provider mangareader.net -manga shingeki-no-kyojin -chapter 100 -path .
 => Download shingeki-no-kyojin chapter 100 and forward into cwd

For the full documentation please refer to:
https://github.com/francoiscolombo/gomangareaderdl`)
	fmt.Println("")
}

func main() {

	fmt.Println("\nWelcome on gomangareaderdl")
	fmt.Println("--------------------------\n")

	fmt.Printf("version %s (%s)\n", versionNumber, versionName)

	if settings.IsSettingsExisting() == false {
		settings.WriteDefaultSettings()
	}

	settings := settings.ReadSettings()

	fmt.Println("- Settings loaded.")
	fmt.Printf("  > Default output path is %s\n  > Default provider is %s\n\n", settings.Config.OutputPath, settings.Config.Provider)

	var params parameters

	flag.StringVar(&params.Command, "command", "???", "command to execute")
	flag.StringVar(&params.Manga, "manga", "???", "manga to download or update")
	flag.IntVar(&params.Chapter, "chapter", -1, "chapter to download")
	flag.IntVar(&params.Next, "next", -1, "overwrite next chapter to download")
	flag.StringVar(&params.Provider, "provider", "???", "Set default provider")
	flag.StringVar(&params.Path, "path", "???", "allow to download manga to another path instead of the default one")
	flag.BoolVar(&params.Force, "force", false, "force download a previously downloaded chapter")
	flag.BoolVar(&params.Silent, "silent", false, "don't display download progress bar")
	flag.StringVar(&params.Output, "output", "???", "set default output path for downloaded mangas")

	flag.Parse()

	// depending the command, right?
	switch params.Command {
	case commandList:
		// list command
		commands.ProcessListCommand(&settings)
	case commandConfig:
		// config command allows the following parameters: output and provider
		commands.ProcessConfigCommand(&settings, params.Output, params.Provider)
	case commandFetch:
		// fetch command allows the following parameters: manga, chapter, provider, path, force and silent
		commands.ProcessFetchCommand(&settings, params.Manga, params.Chapter, params.Provider, params.Path, params.Force, params.Silent)
	case commandUpdate:
		// update command allows the following parameters: manga, provider and next
		commands.ProcessUpdateCommand(&settings, params.Manga, params.Provider, params.Next)
	case commandView:
		// fetch command allows the following parameters: manga, chapter and path
		commands.ProcessViewCommand(&settings, params.Manga, params.Chapter, params.Path)
	default:
		fmt.Println("Sorry my friend, but you didn't give me the good parameters, so I wont be able to help you!")
		fmt.Println("Maybe a little help can be what you really need? Okay, this should be usefull then...")
		fmt.Println()
		usage()
	}

}
