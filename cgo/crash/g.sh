#!/bin/bash

binname=$1
corefile=$2

(gdb ${binname} ${corefile} > ./${corefile}.log 2>&1) <<-GDBEOF
thread apply all bt
GDBEOF

declare -r cgo="runtime.asmcgocall"
declare -r go="runtime.rt0_go"

lineinfo=$(grep ${cgo} ./${corefile}.log | awk -F' ' '{printf $1}')
if [[ "$lineinfo" == "" ]]; then
    lineinfo=$(grep ${go} ./${corefile}.log | awk -F' ' '{printf $1}')
fi
echo "lineinfo:$lineinfo"
if [[ "$lineinfo" == "" ]]; then
    exit 0
fi

end=$(echo "$lineinfo"|cut -d ':' -f1)
lineno=$(echo "$lineinfo"|cut -d '#' -f2)

start=$((end - lineno - 2))
sed -n "$start,$end p" ./${corefile}.log