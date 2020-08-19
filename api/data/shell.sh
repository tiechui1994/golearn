#!/usr/bin/env bash

# mp3 -> wav
for file in $(ls|grep '.mp3') ; do
    in="$file"
    out=${in/%mp3/wav}
    ffmpeg -i ${in} -acodec pcm_s16le -ac 1 -ar 16000 ${out}
done

# pcm -> wav
for file in $(ls|grep '.pcm'); do
   in="$file";
   out=${in/%pcm/wav}
   ffmpeg -i ${in} -f s16le -ar 16000 -ac 1 -acodec pcm_s16le  ${out}
done

mkdir -p amr
mv *.amr amr/
