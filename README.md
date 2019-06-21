# gomangareaderdl

This is a simple cli tool to mass download mangas from several providers, manage history of downloaded chapters, checks if new chapters availables and continue downloads from the last chapter.

## Why?

They are already a lot of cli tools for mass downloading mangas, so why another one?

I wanted to start on golang and this project seems to be a nice starting point!

And the result is actually not so bad. So why not share it, isn't it?

## Features

- download all available chapters for a manga
- allow to easily download new chapters
- maintain an history of the downloaded mangas
- configurable
- nice and simple to use interface

## Install

First, you must have a proper go workspace already set up in order to use this tool.

Then after that, it's as easy that:

    $ git clone https://github.com/francoiscolombo/gomangareaderdl.git
    $ cd gomangareaderdl
    $ go get
    $ go install gomangareaderdl.go

This should produce an executable in your ``${GOPATH}/bin`` directory.

## Usage

    gomangareaderdl: CLI for manga mass download
    
    Usage
     $ gomangareaderdl -command fetch -manga <manga>
    
    Commands list
     fetch     Fetch manga to download
     config    Set defaults
     update    Update subscribed manga
     list      List downloaded manga
    
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
    
    Example
     $ gomangareaderdl -command fetch -provider mangareader.net -manga shingeki-no-kyojin -chapter 100 -path .
     => Download shingeki-no-kyojin chapter 100 and forward into cwd
    
    For the full documentation please refer to:
    https://github.com/francoiscolombo/gomangareaderdl

## How to use it?

### Start by configuring the defaults

If you don't give some optional parameters, the defaults will be used, so it's a good idea to begin by setting up these defaults.

In order to do that, simply use the command ``config`` like this:

    $ gomangareaderdl -command config -output /data/mangas -provider mangareader.net

by issuing this command, you set the *default path* to **/data/mangas** and the *default provider* to **mangareader.net**

### Fetch your favorite mangas

After all it's the main goal of this tool, isn't it?

    $ gomangareaderdl -command fetch -manga shingeki-no-kyojin

This command will start to download the manga *shingeki-no-kyojin* from the chapter 1, and generate cbz for every chapters in the default path **/data/mangas/shingeki-no-kyojin**.

But you can also use a command like this one:

    $ gomangareaderdl -command fetch -provider mangapanda.com -manga shingeki-no-kyojin -chapter 100 -path .

Here you override the default provider and use *mangapanda.com* instead, you start to download from the *chapter 100* and you store your cbz in the directory *./shingeki-no-kyojin*

After downloading the last available chapter, the cli stop. And he keeps in his history the laster downloaded chapter. Which means that the next time you launch this command:

    $ gomangareaderdl -command fetch -manga shingeki-no-kyojin

it will restart from the last downloaded chapter. Nice, no?

### See the history

Once you download a few mangas, you can check your history whith this simple command:

    $ go run gomangareaderdl.go -command list
    
    Welcome on gomangareaderdl
    --------------------------
    
    version 0.2 (atomic thunderbolt)
    
    Loading settings from C:\Users\Francois/.gomangareader.json...
    Successfully Opened settings.json
    - Settings loaded.
      > Default output path is D:\Documents\mangas
      > Default provider is mangareader.net
    
    - List command selected
    +------------------------+--------------+-----------------+
    |          NAME          | LAST CHAPTER |    PROVIDER     |
    +------------------------+--------------+-----------------+
    | onepunch-man           |          159 | mangareader.net |
    | shingeki-no-kyojin     |          119 | mangareader.net |
    | btooom                 |          102 | mangareader.net |
    | the-promised-neverland | 139          | mangareader.net |
    +------------------------+--------------+-----------------+

If a new chapter is available, the manga will be display with another color. So you can easily see what are the new mangas you need to download!

### Rewrite your history

But maybe your last downloaded chapter was corrupted and you want to download it again, but from another provider?

No problems. Just use a command like this:

    $ gomangareaderdl -command update -manga btooom -provider mangapanda.com -next 98

And then your history will change, and you can now download your manga again.

## Supported sites

Currently supported sites:

| site | speed | note |
| --- | --- | --- |
| mangareader.net |	fast | fastest with mangapanda.com |
| mangapanda.com | fast | mangareader rehost |
| mangalife.us | not so fast | wide variety, best formatting for manhwa |
