FROM ubuntu:latest
LABEL VERSION=__version

ARG USER=__config_user
ARG PASSWORD=__config_password
ARG DB_USER=__config_db_user
ARG DB_PASSWORD=__config_db_password
ARG DB_NET=__config_db_net
ARG DB_ADDR=__config_db_addr
ARG DB_NAME=__config_db_name
ARG PORT=__config_port
ARG BUCKET=__config_bucket
ARG ORG=__config_org
ARG TOKEN=__config_token
ARG URL=__config_url
ARG RETENTION=__config_retention

COPY ./frontend/frontend /usr/bin/eeprom-monitoring-server-frontend
COPY ./backend/backend /usr/bin/eeprom-monitoring-server-backend
COPY ./static/ /etc/eeprom-monitoring-server/static/
COPY ./templates/ /etc/eeprom-monitoring-server/templates/
COPY ./config/config.yaml /etc/eeprom-monitoring-server/config.yaml
COPY ./config/mysql.dump /tmp/database.dump

EXPOSE ${PORT}
EXPOSE 8086
