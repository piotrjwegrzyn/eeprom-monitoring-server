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

ENV INFLUX_BUCKET=ems
ENV INFLUX_ORG=eeprom-monitoring-server
ENV INFLUX_TOKEN=v3rY-d1ff1cUlT-t0k3n
ENV INFLUX_HOST=http://localhost
ENV INFLUX_PORT=8086
ENV INFLUX_RETENTION=24h

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
    -t ${INFLUX_TOKEN} \
    -o ${INFLUX_ORG} \
    -b ${INFLUX_BUCKET} \
    -r ${INFLUX_RETENTION} \
    -f -n default --host ${INFLUX_HOST}:${INFLUX_PORT}

ENTRYPOINT /usr/bin/influxd & sleep 10 && service mariadb start & sleep 10 && \
    /usr/bin/ems-frontend -s /etc/ems/static/ -t /etc/ems/templates/ -c /etc/ems/config.yaml & sleep 10 && \
    /usr/bin/ems-backend -c /etc/ems/config.yaml & bash

EXPOSE ${PORT}
EXPOSE ${INFLUX_PORT}
