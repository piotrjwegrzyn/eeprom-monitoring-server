#!/bin/sh

if [ $# -ne 1 ];
then
    exit 1
fi

ip addr show $1 > /dev/null 2>&1
if [ $? -ne 0 ];
then
    exit 1
fi

xxd -p -c 16 $EP_DIR/interfaces/$1 2> /dev/null
