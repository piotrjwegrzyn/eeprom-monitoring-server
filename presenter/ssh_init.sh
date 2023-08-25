#!/bin/sh

echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config
echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
echo "PermitUserEnvironment yes" >> /etc/ssh/sshd_config

mkdir /root/.ssh/
echo "FIN_DIR=$FIN_DIR" >> /root/.ssh/environment
