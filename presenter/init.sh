#!/bin/sh

mkdir $FIN_DIR/interfaces

for interface in $(find $FIN_DIR/eeproms/* -type d | sed 's/\/.*\///')
do
    file=$FIN_DIR/eeproms/$interface/$(ls $FIN_DIR/eeproms/$interface/ | head -1)
    if test -f "$file";
    then
        ln -sf $file $FIN_DIR/interfaces/$interface
    fi
done

ln -s $FIN_DIR/present.sh /usr/bin/show-eeprom
echo "ls --color=never $FIN_DIR/interfaces/ | grep ." > /usr/bin/show-fiber-interfaces
chmod +x /usr/bin/show-fiber-interfaces

# SSH init
echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config
echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
echo "PermitUserEnvironment yes" >> /etc/ssh/sshd_config

mkdir /root/.ssh/
echo "FIN_DIR=$FIN_DIR" >> /root/.ssh/environment
