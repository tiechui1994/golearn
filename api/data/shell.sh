#!/usr/bin/env bash

for file in $(ls|grep '.mp3') ; do
    in="$file"
    out=${in/%mp3/amr}
    ffmpeg -i ${in} -ac 1 -ar 8000 -ab 12.20k ${out}
done

mkdir -p amr
mv *.amr amr/
