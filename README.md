## F1Viewer

Download streamable .m3u8 file for any F1TV VOD.
Play the downloaded files with VLC or MPV player.
With MPV you need to set the flag `--demuxer-lavf-o=protocol_whitelist=[http,https,tls,rtp,tcp,udp,crypto,httpproxy,file]`.

Some files have audio desync issues with VLC.

![alt text](https://i.imgur.com/JHnofPI.png)
 

**USAGE**

    $ go get github.com/rivo/tview
    $ go build
    $ .\F1viewer
    
