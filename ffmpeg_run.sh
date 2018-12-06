#!/bin/bash
#timeout is 10 second. 10000000 (microsecond)
ffmpeg -stimeout 10000000 -rtsp_transport tcp -re -i $1 -vcodec copy -acodec copy -f flv -y $2
