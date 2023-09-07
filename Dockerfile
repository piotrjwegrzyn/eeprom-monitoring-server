FROM ubuntu:latest
LABEL VERSION=latest

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -yq mysql-server

ARG CONFIG=./testdata/ems.yaml
ARG MYSQL_USER=http
ARG MYSQL_PASSWORD=http-password
ARG PORT=80

COPY ./bin/ems-frontend /usr/bin/ems-frontend
COPY ./bin/ems-backend /usr/bin/ems-backend
COPY ./frontend/static/ /etc/ems/static/
COPY ./frontend/templates/ /etc/ems/templates/
COPY ${CONFIG} /etc/ems/config.yaml
COPY ./testdata/mysql.dump /tmp/database.dump
COPY ./bin/influxd /usr/bin/influxd
COPY ./bin/influxc /usr/bin/influx

RUN /usr/sbin/mysqld & sleep 5 && \
    mysql -u root -e "CREATE USER '${MYSQL_USER}'@'localhost' IDENTIFIED BY '${MYSQL_PASSWORD}';" && \
    mysql -u root -e "CREATE DATABASE \`$(grep "dbname" /etc/ems/config.yaml | awk '{printf $2}')\`" && \
    mysql -u root -e "GRANT ALL PRIVILEGES ON \`$(grep "dbname" /etc/ems/config.yaml | awk '{printf $2}')\`.* TO '${MYSQL_USER}'@'localhost';" && \
    mysql -u root -e "FLUSH PRIVILEGES;" && \
    mysql $(grep "dbname" /etc/ems/config.yaml | awk '{printf $2}') < /tmp/database.dump

RUN /usr/bin/influxd & sleep 5 && \
    /usr/bin/influx setup \
    -u $(grep -1 "users" /etc/ems/config.yaml | tail -1 | awk '{printf substr($1, 1, length($1)-1)}') \
    -p $(grep -1 "users" /etc/ems/config.yaml | tail -1 | awk '{printf $2}') \
    -t $(grep "token" /etc/ems/config.yaml | awk '{printf $2}') \
    -o $(grep "org" /etc/ems/config.yaml | awk '{printf $2}') \
    -b $(grep "bucket" /etc/ems/config.yaml | awk '{printf $2}') \
    -r $(grep "retention" /etc/ems/config.yaml | awk '{printf $2}') \
    -f -n default --host http://:8086

ENTRYPOINT /usr/sbin/mysqld & sleep 2 && /usr/bin/influxd & sleep 2 && \
    /usr/bin/ems-frontend -s /etc/ems/static/ -t /etc/ems/templates/ -c /etc/ems/config.yaml & \
    /usr/bin/ems-backend -c /etc/ems/config.yaml & bash

EXPOSE ${PORT}
EXPOSE 8086
