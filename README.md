## F1Viewer

Download streamable .m3u8 file for any F1TV VOD or play directly with [MPV](https://mpv.io/). 
Make sure MPV is added to your `PATH`. Launching MPV can take a few seconds.

You can play the downloaded files with VLC or MPV player. To play them with MPV you need to set the flag `--demuxer-lavf-o=protocol_whitelist=[http,https,tls,rtp,tcp,udp,crypto,httpproxy,file]`.

Some files have audio desync issues with VLC.

![alt text](https://i.imgur.com/J6pgOlQ.png)
 

**USAGE**

    $ go get github.com/rivo/tview
    $ go build
    $ .\F1viewer
    
