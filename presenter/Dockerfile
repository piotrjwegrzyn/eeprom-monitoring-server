FROM ubuntu:latest
LABEL VERSION=latest

ARG EEPROM_ITER=300
ARG SLEEP_TIME=1
ARG EEPROM_SRC

ENV EP_DIR=/usr/src/eeprom-presenter
ENV EEPROM_ITER=${EEPROM_ITER}
ENV SLEEP_TIME=${SLEEP_TIME}
WORKDIR ${EP_DIR}

RUN echo 'root:password' | chpasswd
RUN echo "Welcome to eeprom-presenter based on Ubuntu!\n" > /etc/motd && \
    echo "Type \"show-eeprom <IFACE>\" to get current EEPROM from interface\n" >> /etc/motd && \
    echo "Type \"show-fiber-interfaces\" to list all optical interfaces\n" >> /etc/motd
RUN echo "export EP_DIR=${EP_DIR}" >> /etc/profile && \
    chmod -x /etc/update-motd.d/*

# SSH installation and configuration
RUN apt-get update && apt-get install -y iproute2 openssh-server xxd
RUN echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config && \
    echo "PermitRootLogin yes" >> /etc/ssh/sshd_config && \
    echo "PermitUserEnvironment yes" >> /etc/ssh/sshd_config && \
    mkdir /root/.ssh/ && \
    echo "EP_DIR=$EP_DIR" >> /root/.ssh/environment

# Setting up program
COPY presenter/present.sh presenter/update.sh ${EP_DIR}
COPY ${EEPROM_SRC} ${EP_DIR}/eeprom/

RUN mkdir ${EP_DIR}/interfaces && \
    for interface in $(find ${EP_DIR}/eeprom/* -type d | sed 's/\/.*\///'); \
    do \
        file=${EP_DIR}/eeprom/$interface/$(ls ${EP_DIR}/eeprom/$interface/ | head -1); \
        if test -f "$file"; \
        then \
            ln -sf $file ${EP_DIR}/interfaces/$interface; \
        fi; \
    done

RUN ln -s ${EP_DIR}/present.sh /usr/bin/show-eeprom && \
    echo "ls --color=never ${EP_DIR}/interfaces/ | grep ." > /usr/bin/show-fiber-interfaces && \
    chmod +x /usr/bin/show-fiber-interfaces

WORKDIR /root
ENTRYPOINT service ssh start > /dev/null 2>&1 && ${EP_DIR}/update.sh & /bin/sh

EXPOSE 22
