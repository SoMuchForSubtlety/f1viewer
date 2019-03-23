## F1Viewer

Stream any F1TV VOD with [MPV](https://mpv.io/) or download the corresponding .m3u8 file. 
Make sure MPV is added to your `PATH`.

Now also supports live streams (except the main feed).

You can play the downloaded files with VLC or MPV player. To play them with MPV you need to set the flag `--demuxer-lavf-o=protocol_whitelist=[http,https,tls,rtp,tcp,udp,crypto,httpproxy,file]`.

Some files have audio desync issues with VLC.

![alt text](https://i.imgur.com/K8yCkib.png)
 

**USAGE**

    $ cp sample-config.json config.json
    $ go get github.com/rivo/tview
    $ go build
    $ .\F1viewer
    

**FLAGS**

    -d
shows a debug window

    -vlc
enables the VLC http stream option
