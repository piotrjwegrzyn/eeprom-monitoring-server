FROM golang:1.23.3 AS build
WORKDIR /go/src
COPY . .
ENV CGO_ENABLED=0
RUN go build -C frontend -ldflags="-w -s" -trimpath -o /ems-frontend
RUN go build -C backend -ldflags="-w -s" -trimpath -o /ems-backend

FROM golang:1.23.3-bookworm AS aio
LABEL VERSION=latest

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -yq mariadb-server
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

ARG CONFIG=./testdata/ems.yaml
ARG PORT=80
ENV DB_NAME=mysql
ENV DB_USER=http
ENV DB_PASSWORD=http-password
ENV DB_HOST=localhost
ENV DB_PORT=3306

COPY --from=build /ems-frontend /usr/bin/ems-frontend
COPY --from=build /ems-backend /usr/bin/ems-backend
COPY ./bin/influxd /usr/bin/influxd
COPY ./bin/influxc /usr/bin/influx
COPY ./frontend/static/ /etc/ems/static/
COPY ./frontend/templates/ /etc/ems/templates/
COPY ./storage/sqlc/migrations/ /etc/ems/migrations/
COPY ${CONFIG} /etc/ems/config.yaml

RUN service mariadb start & sleep 10 && \
    mysql -u root -e "CREATE USER '${DB_USER}'@'${DB_HOST}' IDENTIFIED BY '${DB_PASSWORD}';" && \
    mysql -u root -e "GRANT ALL PRIVILEGES ON *.* TO '${DB_USER}'@'${DB_HOST}'; FLUSH PRIVILEGES;" && \
    goose -dir "/etc/ems/migrations" -table db_version mysql "${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=True" up

RUN /usr/bin/influxd & sleep 5 && \
    /usr/bin/influx setup \
    -u $(grep -1 "users" /etc/ems/config.yaml | tail -1 | awk '{printf substr($1, 1, length($1)-1)}') \
    -p $(grep -1 "users" /etc/ems/config.yaml | tail -1 | awk '{printf $2}') \
    -t $(grep "token" /etc/ems/config.yaml | awk '{printf $2}') \
    -o $(grep "org" /etc/ems/config.yaml | awk '{printf $2}') \
    -b $(grep "bucket" /etc/ems/config.yaml | awk '{printf $2}') \
    -r $(grep "retention" /etc/ems/config.yaml | awk '{printf $2}') \
    -f -n default --host http://:8086

ENTRYPOINT service mariadb start & sleep 2 && /usr/bin/influxd & sleep 2 && \
    /usr/bin/ems-frontend -s /etc/ems/static/ -t /etc/ems/templates/ -c /etc/ems/config.yaml & \
    /usr/bin/ems-backend -c /etc/ems/config.yaml & bash

EXPOSE ${PORT}
EXPOSE 8086
