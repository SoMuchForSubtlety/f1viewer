[![Go Report Card](https://goreportcard.com/badge/github.com/SoMuchForSubtlety/f1viewer)](https://goreportcard.com/report/github.com/SoMuchForSubtlety/f1viewer)
![](https://github.com/SoMuchForSubtlety/f1viewer/workflows/Test/badge.svg)

# f1viewer

![preview image](https://user-images.githubusercontent.com/15961647/107859733-c6a8a900-6e3b-11eb-82b8-5b1ee0a16297.png)

## Table of Contents

* [Installation](#Installation)
* [FAQ](#Faq)
* [Config](#Config)
* [Custom Commands](#Custom-commands)
* [Multi Commands](#Multi-commands)
* [Live Session Hooks](#Live-session-hooks)
* [Key Bindings](#Key-bindings)
* [Logs](#Logs)
* [Credentials](#Credentials)

## Installation

**Note:** You also need a compatible player installed, you can find a list [here](https://github.com/SoMuchForSubtlety/f1viewer/wiki/Players-Supported-by-Default).

## compile form source
Install the go compiler, then run the following commands
```bash
git clone https://github.com/SoMuchForSubtlety/f1viewer && cd f1viewer
go build .
```

### Windows
* Download [the latest release directly](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)
* Or install with [chocolatey](https://chocolatey.org/packages/f1viewer/) 

### macOS
* You can install f1viewer with Homebrew (recommended)
	```bash
	brew tap SoMuchForSubtlety/tap
	brew install SoMuchForSubtlety/tap/f1viewer
	```
* Or [download the binary directly](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)

### Debian and Ubuntu
Download the latest release `.deb` [file](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)

### Fedora, openSUSE, CentOS
 * Install from the f1viewer [copr repo](https://copr.fedorainfracloud.org/coprs/somuchforsubtlety/f1viewer/)

   ```bash
   sudo dnf install dnf-plugins-core
   sudo dnf copr enable somuchforsubtlety/f1viewer
   sudo dnf install f1viewer
   ```

 * Or download the latest release `.rpm` [file](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)

### Arch
Install the f1viewer [AUR package](https://aur.archlinux.org/packages/f1viewer/).

### Any other Linux distribution
* Download the binary [directly](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)
* Or install it with [Homebrew](https://docs.brew.sh/Homebrew-on-Linux) as described in the [macOS](#macOS) section.

## FAQ
#### why is there a login, what credentials should I use
You need an F1TV account created with an IP in a country that has F1TV pro. Use your F1TV account email and password to log in. You can use the tab key to navigate the login form.
#### when I try to play something I get a 4xx error
You need to be logged in and in a country that has F1TV pro. If you get the error but think your account should be able to play the selected content please open an issue.
#### f1viewer is not showing a live session / loading very slowly
This can happen if the F1TV servers are overloaded. There is nothing I can do to fix this.
Start your stream as soon as possible at the start of the session and you can usually avoid this.
#### The player starts but then has some issue / error
Please make sure you are using the latest version of the player. If you use Windows please download MPV from [here](https://sourceforge.net/projects/mpv-player-windows/files/). Generally once an external program is started f1viewer is done and you should consult the external program's documentation for troubleshooting. 
#### No players are detected
Players need to be in your PATH environment variable to be detected by f1viewer.

## Config
When you first start f1viewer a boilerplate config is automatically generated. On Widows systems it's located in `%AppData%\Roaming\f1viewer`, on macOS in `$HOME/Library/Application Support/f1viewer` and on Linux in `$XDG_CONFIG_HOME/f1viewer` or `$HOME/.config/f1viewer`. You can access it quickly by running `f1viewer -config`.

## Custom Commands
You can execute custom commands, for example to launch a different player. These are set in the config under `custom_playback_options` in the config file. You can add as many as you want.
```toml
[[custom_playback_options]]
  command = ["ffmpeg", "-hide_banner", "-loglevel", "error", "-i", "$url", "-c", "copy", "-f", "mp4", "$title.mp4"]
  proxy   = true
  title   = "Download as mp4"
```

`title` is the title. It will appear next to the standard `Play with MPV` and `Copy URL to clipboard`.

`command` is where your command goes. It is saved as a list of args like in the examples above. Every argument should be a separate string! The following would be incorrect! `["ffmpeg", "-i $url", "-c copy", "$title.mp4"]`

`proxy` sends http requests through a proxy if they require cookies. This is useful for commands that use ffmpeg (and by extension mpv).

There are several placeholder variables you can use that will be replaced by f1viewer.

 - `$url`: the content's URL
 - `$category`: the content's category (eg. "Documentary")
 - `$season`: the season's year (eg. "2021")
 - `$event`: the event (eg. "Belgian Grand Prix")
 - `$session`: the session (eg. "F1 Practice 3")
 - `$perspective`: the perspective (eg. "F1 Live", "Kimi Räikkönen", etc.)
 - `$title`: the conten's title as reported by F1TV
 - `$filename`: the same as title, but with illegal characters removed
 - `$series`: "Formula 1", "Formula 2", etc.
 - `$country`: the country an event is held in
 - `$circuit`: the circuirt and event is held at
 - `$time`: the time of the session in RFC3339 format (`$year`, `$month`, `$day`, `$hour` and `$minute` are also available)
 - `$date`: the date of the session in ISO 8601 format
 - `$ordinal`: the ordinal numer of the event
 - `$episodenumber`: the episode number as reported by F1TV
 - `$json`: all metadata fields and the full source metadata from F1TV
 - `$lang`: the preferred languages as a comma separated list

If you have ideas for more variables feel free to open an issue.

**Tip**: To get Windows commands like `echo`, `dir`, etc. to work, you'll need to prepend them with `"cmd", "/C"`, so for example `["echo", "hello"]` turns into `["cmd", "/C", "echo", "hello"]`

## Multi Commands
To make it easy to load the same feeds with the same commands every time, you can map multiple commands to one action. The `match_title` variable will be used to match the session feeds (it also allows regex). For example, if `match_title` is `Lando Norris`, it will load any feed with that name, with the given command.
You can specify commands directly with `command`, or reference one of your [custom commands](#custom-command) titles with `command_key`.

For an explanation on the `command` variable, see [Custom Commands](#custom-commands)

```toml
[[multi_commands]]
  title = "Open F1 Live and HAM onboard"

  [[multi_commands.targets]]
    command     = ["mpv", "$url", "--alang=$lang"] # define a command to execute
    match_title = "F1 Live"

  [[multi_commands.targets]]
    command_key = "custom mpv"      # you can also reference previously defined custom commands
    match_title = "Lewis [a-zA-Z]+" # regex is also supported
```

## Live Session Hooks
Live session hooks work like multi commands, but they are automatically started when a new live session is detected.

```toml
[[live_session_hooks]]
  title = "Open Pit Lane and Data Channel"

  [[live_session_hooks.targets]]
    command     = ["mpv", "$url", "--alang=$lang", "--quiet"] # define a command to execute
    match_title = "Pit Lane"

  [[live_session_hooks.targets]]
    command_key = "custom mpv"      # you can also reference previously defined custom commands
    match_title = "Data Channel"
```

## Key Bindings
* arrow keys or `h`, `j`, `k`, `l`.  
* `tab` to cycle through the login form fields
* enter to select / confirm
* `q` to quit

## Logs
By default f1viewer saves all info and error messages to log files. Under Windows and macOS they are save in the same directory as the config file, on Linux they are saved to `$HOME/.local/share/f1viewer/`. You can access them quickly by running `f1viewer -logs`.
Saving logs can also be turned off in the config.

## Credentials
Your login credentials for F1TV are not saved in the config file. On macOS they are stored in the keychain and on Windows the credential store is used. If you're using Linux, where they are saved depends on your distro. Generally [Pass](https://www.passwordstore.org/), [Secret Service](https://specifications.freedesktop.org/secret-service/latest/) / [GNOME Keyring](https://wiki.gnome.org/Projects/GnomeKeyring) and KWallet are supported.
If it does not work on your distro or you encounter any problems please open an issue.