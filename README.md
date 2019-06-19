# F1Viewer

Watch any F1TV VOD or live stream. 

![preview image](https://i.imgur.com/DwHnnK9.png)

## Table of Contents   
* [Usage](#usage)
* [Flags](#flags)
* [FAQ](#faq)
* [Config](#config)
* [Custom Commands](#custom-commands)

## Usage

 1. **get F1Viewer** 

	download [pre-compiled binaries](https://github.com/SoMuchForSubtlety/F1viewer/releases/)

	**or**

	build it yourself
	
	    $ git clone https://github.com/SoMuchForSubtlety/F1viewer/
	    $ cd F1Viewer
	    $ go build
	    
 2. **Download MPV**

	Download it from [here](https://mpv.io/installation/) (Windows users please download from [here](https://sourceforge.net/projects/mpv-player-windows/files/)) and either put it in the same folder as the  F1Viewer binary or add it to your PATH environment variable.
	
	(You can also use other players, see [Custom Commands](#custom-commands)) 

## Flags

    -d
shows debug information

## FAQ
#### F1Viewer is not showing a live session / loading very slowly
This can happen if the F1TV servers are overloaded. There is nothing I can do to fix this.
Start your stream as soon as possible at the start of the session and you can usually avoid this. 
#### I downloaded a .m3u8 file but can't play it
F1TV now requires a cookie to be set to open the links in a .m3u8 file. You can get the cookie with `$cookie` though a custom command or  completely circumvent this issue by using the url to the .m3u8 instead (custom command `$url`).
If you need a local file put the link in a .strm file.
#### MPV is opening but I'm not getting audio
Please make sure you are using the latest version of MPV. If you use  Windows  please download it from [here](https://sourceforge.net/projects/mpv-player-windows/files/).

## Config
The config is is optional. It is used to set a preferred audio language and custom commands. It can also be used to stop checking for updates. 
It should look like this.

    {
        "preferred_language": "en",
        "check_updates": true,
        "custom_playback_options": [
            {
                "title": "Play with MPV custom",
                "commands": [
                    ["mpv", "$url", "--alang=de"]
                ],
                "watchphrase": "Video",
                "command_to_watch": 0
            }
        ]
    }

Save `sample-config.json` as `config.json` in the same Folder as the F1Viewer binary and edit it so it fits your needs.

## Custom Commands
You can execute custom commands, for example to launch a different player. These are set in the config under `custom_playback_options`. You can add as many as you want. 

`title` is the title. It will appear next to the standard `Play with MPV` and `Download .m3u8`.

`commands` is where your custom command goes. There can be one or more. 
Commands are saved as a list of args, like `["mpv", "$url", "--alang=de"]`.  
`$url` will be replaced with the playback URL.  
`$file` will be replaced with the path to a local copy of the .m3u8 file.  
`$cookie` will be replaced with the cookie you get by downloading a .m3u8 file with `$file`, this needs to be set or you will  get 403 errors when you try and play the file.

With `concurrent` you can set whether one command should finish before the next one is executed, or they all launch simultaneously. It defaults to false and is only needed if there is more than one command. (see `sample-config.json` for example)

`watchphrase` is optional. it is used to play a `loading...` animation. 
F1Viewer can parse the output of your command and stop the animation once the `watchphrase` is found. This can be useful if your command takes a while to execute.

`command_to_watch` belongs to `watchphrase`. It defines what command to parse if there are multiple. It is indexed at 0 so if you only have 1 command, `command_to_watch` should be `0`.

If `command_to_watch` is out of range or `watchphrase` is an empty string, the loading animation will be skipped.
