FROM ubuntu:latest
LABEL VERSION=__version

ARG CONFIG=__config_file
ARG USER=__config_user
ARG PASSWORD=__config_password
ARG DB_USER=__config_db_user
ARG DB_PASSWORD=__config_db_password
ARG DB_NAME=__config_db_name
ARG PORT=__config_port
ARG BUCKET=__config_bucket
ARG ORG=__config_org
ARG TOKEN=__config_token
ARG RETENTION=__config_retention

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -yq mysql-server

COPY ./frontend/frontend /usr/bin/eeprom-monitoring-server-frontend
COPY ./backend/backend /usr/bin/eeprom-monitoring-server-backend
COPY ./static/ /etc/eeprom-monitoring-server/static/
COPY ./templates/ /etc/eeprom-monitoring-server/templates/
COPY ./${CONFIG} /etc/eeprom-monitoring-server/config.yaml
COPY ./config/mysql.dump /tmp/database.dump
COPY ./influx/influxd /usr/bin/influxd
COPY ./influx/influx /usr/bin/influx

RUN /usr/sbin/mysqld & sleep 5 && \
    mysql -u root -e "CREATE USER '${DB_USER}'@'localhost' IDENTIFIED BY '${DB_PASSWORD}';" && \
    mysql -u root -e "CREATE DATABASE \`${DB_NAME}\`" && \
    mysql -u root -e "GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'localhost';" && \
    mysql -u root -e "FLUSH PRIVILEGES;" && \
    mysql ${DB_NAME} < /tmp/database.dump
RUN /usr/bin/influxd & sleep 5 && \
    /usr/bin/influx setup -u ${USER} -p ${PASSWORD} -t ${TOKEN} -o ${ORG} -b ${BUCKET} -r ${RETENTION} -f -n default --host http://:8086

ENTRYPOINT /usr/sbin/mysqld & sleep 2 && /usr/bin/influxd & sleep 2 && \
    /usr/bin/eeprom-monitoring-server-frontend -static /etc/eeprom-monitoring-server/static/ -templates /etc/eeprom-monitoring-server/templates/ -config /etc/eeprom-monitoring-server/config.yaml & \
    /usr/bin/eeprom-monitoring-server-backend -config /etc/eeprom-monitoring-server/config.yaml & bash

EXPOSE ${PORT}
EXPOSE 8086
