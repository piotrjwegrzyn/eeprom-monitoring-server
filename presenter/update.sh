#!/bin/sh

for iteration in $(seq 0 $EEPROM_ITER);
do
    for interface in $(find $EP_DIR/eeprom/* -type d | sed 's/\/.*\///')
    do
        file=$(printf $EP_DIR/eeprom/$interface/$interface-%09d $iteration)
        if test -f "$file";
        then
            ln -sf $file $EP_DIR/interfaces/$interface
        fi
    done
    sleep $SLEEP_TIME
done
