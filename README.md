

## F1Viewer

Stream any F1TV VOD with MPV or download the corresponding .m3u8 file. 

Now also supports live streams.

![alt text](https://i.imgur.com/K8yCkib.png)
 

### USAGE

 1. **install F1Viewer** 

	build it yourself
	
	    $ git clone https://github.com/SoMuchForSubtlety/F1viewer/
	    $ cd F1Viewer
	    $ cp sample-config.json config.json
	    $ go get github.com/rivo/tview
	    $ go build

	    
	**or**
    
	or download pre-compiled binaries
https://github.com/SoMuchForSubtlety/F1viewer/releases/

 2. **Download MPV**

	Download it from [here](https://mpv.io/installation/) and either put it in the same folder as the  F1Viewer binary or add it to your PATH environment variable.


    

### FLAGS

    -d
shows debug information

### CONFIG
The config is is optional. It is used to set a preferred audio language and custom commands.
The sample config looks like this.

    {
        "preferred_language": "en",
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

Save it as `config.json` in the same Folder as the F1Viewer binary 

### CUSTOM COMMANDS
You can execute custom commands, for example to launch a different player. These are set in the config under `custom_playback_options`. You can add as many as you want. 

`title` is the title. It will appear next to the standard `Play with MPV` and `Download .m3u8`.

`commands` is where your custom command goes. There can be one or more. 
Commands are saved as a list of args, like `["mpv", "$url", "--alang=de"]`.  
`$url` will be replaced with the playback URL.  
`$file` will be replaced with the path to a local copy of the .m3u8 file.

`watchphrase` is optional. it is used to play a `loading...` animation. 
F1Viewer can parse the output of your command and stop the animation once the `watchphrase` is found. This can be useful if your command takes a while to execute.

`command_to_watch` belongs to `watchphrase`. It defines what command to parse if there are multiple. It is indexed at 0 so if you only have 1 command, `command_to_watch` should be `0`.

If `command_to_watch` is out of range or `watchphrase` is an empty string, the loading animation will be skipped.
