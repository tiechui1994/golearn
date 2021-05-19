#!/usr/bin/env bash

declare  utils=''
declare -A main=()

for i in $(ls) ; do
    if [[ -f ${i} && ${i} =~ .go$ ]]; then
        if [[ ${i} =~ _test.go$ ]]; then
            continue
        fi

        if [[ $(grep -E '^func[[:space:]]+main()' ${i}) ]]; then
            name=${i%.go}
            main[$name]="$i"
        else
            utils="$utils $i"
        fi
    fi
done


#util=init.go
#
#cleam:
#	rm -rf gitee
#
#gitee: cleam
#	go build -o gitee -v -ldflags '-w' $(util) gitee.go
#
#drive: cleam
#	go build -o drive -v -ldflags '-w' $(util) drive.go
#
#all: gitee drive


echo "util=$utils" > /tmp/Makefile
echo >> /tmp/Makefile

# clean
echo "clean:" >> /tmp/Makefile
for name in ${!main[@]}; do
    echo -e "\trm -rf $name" >> /tmp/Makefile
done
echo >> /tmp/Makefile

# build
for name in ${!main[@]}; do
    echo "$name: clean"  >> /tmp/Makefile
    echo -e "\tgo build -o $name -v -ldflags '-w'"' $(util) ' "${main[$name]}" >> /tmp/Makefile
    echo >> /tmp/Makefile
done

# all
echo "all: ${!main[@]}"  >> /tmp/Makefile
echo >> /tmp/Makefile

mv /tmp/Makefile Makefile
