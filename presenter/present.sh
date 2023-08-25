#!/bin/sh

if [ $# -ne 1 ];
then
    exit 1
fi

ip addr show $1 &> /dev/null
if [ $? -ne 0 ];
then
    exit 1
fi

xxd -p -c 16 $FIN_DIR/interfaces/$1 2> /dev/null
# Alternatively:
# xxd $FIN_DIR/interfaces/$1
# hexdump -v -C $FIN_DIR/interfaces/$1
# hexdump -v $FIN_DIR/interfaces/$1
