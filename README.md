[![Go Report Card](https://goreportcard.com/badge/github.com/SoMuchForSubtlety/F1viewer)](https://goreportcard.com/report/github.com/SoMuchForSubtlety/F1viewer)
![](https://github.com/SoMuchForSubtlety/F1viewer/workflows/Test/badge.svg)

# F1viewer

### F1viewer requires an F1TV pro account to stream VODs!

![preview image](https://i.imgur.com/ik2ZRw5.png)

## Table of Contents

* [Usage](#usage)
* [FAQ](#faq)
* [Config](#config)
* [Custom Commands](#custom-commands)
* [Logs](#logs)

## Usage

 1. **get F1viewer** 

	download [pre-compiled binaries](https://github.com/SoMuchForSubtlety/F1viewer/releases/)

	**or**

	build it yourself
	
	    $ git clone https://github.com/SoMuchForSubtlety/F1viewer/
	    $ cd F1viewer
	    $ go build
	    
 2. **Download MPV**

	Download it from [here](https://mpv.io/installation/) (Windows users please download from [here](https://sourceforge.net/projects/mpv-player-windows/files/)) and either put it in the same folder as the  F1viewer binary or add it to your PATH environment variable.
	
	(You can also use other players, see [Custom Commands](#custom-commands)) 

## FAQ
#### F1viewer is not showing a live session / loading very slowly
This can happen if the F1TV servers are overloaded. There is nothing I can do to fix this.
Start your stream as soon as possible at the start of the session and you can usually avoid this. 
#### MPV is opening but I'm not getting audio
Please make sure you are using the latest version of MPV. If you use Windows please download it from [here](https://sourceforge.net/projects/mpv-player-windows/files/).

## Config
When you first start F1viewer a boilerplate config is automatically generated. On Widows systems it's located in `%AppData%\Roaming\F1viewer`, on macOS in `$HOME/Library/Application Support/F1viewer` and on Linux in `$XDG_CONFIG_HOME/F1viewer` or `$HOME/.config/F1viewer`.

The default config looks like this
```json
{
	"username": "",
	"password": "",
	"live_retry_timeout": 60,
	"preferred_language": "en",
	"check_updates": true,
	"save_logs": true,
	"log_location": "",
	"custom_playback_options": null,
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
 - `username` your F1TV username
 - `password` your F1TV password
 - `live_retry_timeout` is the interval F1viewer looks for a live F1TV session seconds
 - `preferred_language` is the language MPV is started with, so the correct audio track gets selected
 - `check_updates` determines if F1TV should check GitHub for new versions
 - `save_logs` determines if logs should be saved
 - `log_location` can be used to set a custom log output folder
 - `custom_playback_options` can be used to set custom commands, see  [Custom Commands](#custom-commands)  for more info
 - `horizontal_layout` can be used to switch the orientation from vertical to horizontal
 - `theme` can be used to set custom colors for various UI elements. Please use standard hex RGB values in the format `#FFFFFF` or `FFFFFF`.
 - `tree_ratio` and `output_ratio` can adjust the UI ratio. The values need to be intergers >= 1.

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

There are several placeholder variables you can use that will be replaced by F1viewer.

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

## Logs
By default F1viewer saves all info and error messages to log files. Under Windows and macOS they are save in the same directory as the config file, on Linux they are saved to `$HOME/.local/share/F1viewer/`. 
The log folder can be changed in the config. Logs can also be turned off completely.
