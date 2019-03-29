## F1Viewer

Stream any F1TV VOD with MPV or download the corresponding .m3u8 file. 

Now also supports live streams.

![screenshot](https://i.imgur.com/K8yCkib.png)
 
### USAGE

 1. **Install F1Viewer** 

	build it yourself
	
	    $ git clone https://github.com/SoMuchForSubtlety/F1viewer/
	    $ cd F1Viewer
	    $ cp sample-config.json config.json
	    $ go get github.com/rivo/tview
	    $ go build

	    
	**or**
    
	download [pre-compiled binaries](https://github.com/SoMuchForSubtlety/F1viewer/releases/)


 2. **Download MPV**

	Download it from [here](https://mpv.io/installation/) and either put it in the same folder as the  F1Viewer binary or add it to your PATH environment variable.

### FLAGS

    -d
shows debug information

    -vlc
enables the VLC http stream option, requires a config with the correct parameters set.

### CONFIG
The config is optional. It is used to set a preferred audio language and VLC streaming credentials.
The sample config looks like this.

    {
        "preferred_language": "en",
        "vlc_telnet_port": "4212",
        "vlc_telnet_pass": "admin"
    }
Save it as `config.json` in the same Folder as the F1Viewer binary 
