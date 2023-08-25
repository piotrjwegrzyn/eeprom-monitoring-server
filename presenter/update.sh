#!/bin/sh

for iteration in $(seq 1 $ITERATIONS);
do
    for interface in $(find $FIN_DIR/eeproms/* -type d | sed 's/\/.*\///')
    do
        file=$(printf $FIN_DIR/eeproms/$interface/$interface-%09d $iteration)
        if test -f "$file";
        then
            ln -sf $file $FIN_DIR/interfaces/$interface
        fi
    done
    sleep $SLEEP_TIME
done
