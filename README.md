[![Go Report Card](https://goreportcard.com/badge/github.com/SoMuchForSubtlety/f1viewer)](https://goreportcard.com/report/github.com/SoMuchForSubtlety/f1viewer)
![](https://github.com/SoMuchForSubtlety/f1viewer/workflows/Test/badge.svg)

# f1viewer

![preview image](https://i.imgur.com/ik2ZRw5.png)

## Table of Contents

* [Installation](#Installation)
* [FAQ](#Faq)
* [Config](#Config)
* [Custom Commands](#Custom-commands)
* [Logs](#Logs)
* [Credentials](#Credentials)

## Installation

### Windows
1. Download and [the latest release for Windows](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)
2. Download [MPV](https://mpv.io/installation/)
**Note:** MPV needs to be in the same folder as f1viewer or added to your PATH

### macOS
* You can install f1viewer with Homebrew
	```bash
	brew tap SoMuchForSubtlety/tap
	brew install SoMuchForSubtlety/tap/f1viewer
	```
* Or [download the binary directly](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)

### Debian and Ubuntu
Download the latest release `.deb` [file](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest) and install it.

### Fedora and openSUSE
Download the latest release `.rpm` [file](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest) and install it.

### Arch
Install the f1viewer [AUR package](https://aur.archlinux.org/packages/f1viewer/).

### Any other Linux distribution
* Download the binary [directly](https://github.com/SoMuchForSubtlety/f1viewer/releases/latest)
* Install it with [Homebrew](https://docs.brew.sh/Homebrew-on-Linux) as described in the [macOS](#macOS) section.
* build it from source
  ```bash
  go get github.com/SoMuchForSubtlety/f1viewer
  ```

## FAQ
#### why is there a login, what credentials should I use
You need an F1TV account to play VODs. Use your F1TV account email and password to log in. You can use the tab key to navigate the login form.
#### when I try to play something I get a 4xx error
You need an F1TV Access or Pro account to watch replays and live streams respectively. If you get the error but think your account should be able to play the selected content please open an issue.
#### f1viewer is not showing a live session / loading very slowly
This can happen if the F1TV servers are overloaded. There is nothing I can do to fix this.
Start your stream as soon as possible at the start of the session and you can usually avoid this.
#### MPV is opening but I'm not getting audio
Please make sure you are using the latest version of MPV. If you use Windows please download it from [here](https://sourceforge.net/projects/mpv-player-windows/files/).

## Config
When you first start f1viewer a boilerplate config is automatically generated. On Widows systems it's located in `%AppData%\Roaming\f1viewer`, on macOS in `$HOME/Library/Application Support/f1viewer` and on Linux in `$XDG_CONFIG_HOME/f1viewer` or `$HOME/.config/f1viewer`.

The default config looks like this
```json
{
	"live_retry_timeout": 60,
	"preferred_language": "en",
	"check_updates": true,
	"save_logs": true,
	"log_location": "",
	"custom_playback_options": [],
	"favorites_playback_options": [],
	"horizontal_layout": false,
	"tree_ratio": 1,
	"output_ratio": 1,
	"theme": {
		"background_color": "",
		"border_color": "",
		"category_node_color": "",
		"folder_node_color": "",
		"item_node_color": "",
		"action_node_color": "",
		"loading_color": "",
		"live_color": "",
		"update_color": "",
		"no_content_color": "",
		"info_color": "",
		"error_color": "",
		"terminal_accent_color": "",
		"terminal_text_color": ""
	}
}
```
 - `live_retry_timeout` is the interval f1viewer looks for a live F1TV session seconds
 - `preferred_language` is the language MPV is started with, so the correct audio track gets selected
 - `check_updates` determines if F1TV should check GitHub for new versions
 - `save_logs` determines if logs should be saved
 - `log_location` can be used to set a custom log output folder
 - `custom_playback_options` can be used to set custom commands, see  [Custom Commands](#custom-commands)  for more info
 - `favorites_playback_options` can be used to load a set of feeds automatically, see [Favorites Commands](#favorite-commands) for more info
 - `horizontal_layout` can be used to switch the orientation from vertical to horizontal
 - `theme` can be used to set custom colors for various UI elements. Please use standard hex RGB values in the format `#FFFFFF` or `FFFFFF`.
 - `tree_ratio` and `output_ratio` can adjust the UI ratio. The values need to be integers >= 1.

## Custom Commands
You can execute custom commands, for example to launch a different player. These are set in the config under `custom_playback_options`. You can add as many as you want.
```json
"custom_playback_options": [
	{
		"title": "download with ffmpeg",
		"command": ["ffmpeg", "-i", "$url", "-c", "copy", "$title.mp4"]
	},
	{
		"title": "create .strm file",
		"command": ["echo", "$url", ">$title.strm"]
	}
]
```

`title` is the title. It will appear next to the standard `Play with MPV` and `Copy URL to clipboard`.

`command` is where your command goes. It is saved as a list of args like in the examples above. Every argument should be a separate string! The following would be incorrect! `["ffmpeg", "-i $url", "-c copy", "$title.mp4"]`

There are several placeholder variables you can use that will be replaced by f1viewer.

 - `$url`: the content's URL
 - `$category`: the content's category (eg. "Documentary")
 - `$season`: the season name (eg. "2019 Formula 1 World Championship")
 - `$event`: the event (eg. "Belgian Grand Prix")
 - `$session`: the session (eg. "F1 Practice 3")
 - `$perspective`: the perspective (eg. "Main Feed", "Kimi Räikkönen", etc.)
 - `$episode`: the name of the episode (eg. "Chasing The Dream - Episode 1")
 - `$title`: a formatted combination of `$category`,  `$season`, `$event` , `$session`, `$perspective` and `$episode` depending on what is available for the given content. (eg. "2019 Formula 1 World Championship - Singapore Grand Prix - Race - Main Feed")

**Note**: `$title` has illegal characters removed so it can be used as a filename, the other variables are left unmodified.

If you have ideas for more variables feel free to open an issue.

**Tip**: To get Windows commands like `echo`, `dir`, etc. to work, you'll need to prepend them with `"cmd", "/C"`, so for example `["echo", "hello"]` turns into `["cmd", "/C", "echo", "hello"]`

## Favorite Commands
To make it easy to load the same feeds with the same commands every time, you can load your favorite commands. The `title` variable will be used to match the session feeds. For example, if your `title` is `Max Verstappen`, it will load any feed with that name, with the given command.

For an explanation on the `command` variable, see [Custom Commands](#custom-commands)

```json
"favorites_playback_options": [
	{
		"title": "Max Verstappen",
		"command": ["mpv", "--alang=en", "--start=0", "--quiet", "--no-border", "--title=$title", "--autofit-larger=1240", "-geometry=0:0", "$url"]
	},
	{
		"title": "Lewis Hamilton",
		"command": ["mpv", "--alang=en", "--start=0", "--quiet", "--no-border", "--title=$title", "--autofit-larger=1240", "-geometry=100%:0", "$url"]
	}
]
```

## Logs
By default f1viewer saves all info and error messages to log files. Under Windows and macOS they are save in the same directory as the config file, on Linux they are saved to `$HOME/.local/share/f1viewer/`.
The log folder can be changed in the config. Logs can also be turned off completely.

## Credentials
Your login credentials for F1TV are not saved in the config file. On macOS they are stored in the keychain and on Windows the credential store is used. If you're using Linux, where they are saved depends on your distro. Generally [Pass](https://www.passwordstore.org/), [Secret Service](https://specifications.freedesktop.org/secret-service/latest/) / [GNOME Keyring](https://wiki.gnome.org/Projects/GnomeKeyring) and KWallet are supported.
If it does not work on your distro or you encounter any problems please open an issue.
