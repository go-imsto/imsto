FROM golang:1.13-alpine as build
MAINTAINER Eagle Liut <eagle@dantin.me>

RUN cat /etc/apk/repositories \
  && sed -ri "s/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/" /etc/apk/repositories \
  && cat /etc/apk/repositories \
  && apk add --update \
  build-base \
  && rm -rf /var/cache/apk/*

ENV GO111MODULE=on GOPROXY=https://goproxy.io,direct ROOF=github.com/go-imsto/imsto
WORKDIR /go/src/$ROOF
ADD . /go/src/$ROOF

RUN go version && go env \
  && go get github.com/ddollar/forego \
  && export LDFLAGS="-X ${ROOF}/cmd.Version=$(date '+%Y%m%d')" \
  && env \
  && GOOS=linux go install -ldflags "${LDFLAGS} -s -w" . \
  && GOOS=linux go install -ldflags "${LDFLAGS} -s -w" ./apps/imsto-admin \
  && echo "build done"


FROM alpine:3.10

ENV PGDATA=/var/lib/postgresql/data \
    PG_INITDB_OPTS="--encoding=UTF8 --locale=en_US.UTF-8 --auth=trust" \
    PG_LISTEN=localhost \
    PGHOST="localhost" DB_NAME=imsto DB_USER=imsto DB_PASS=mypassword \
    IMSTO_META_DSN='postgres://imsto@localhost/imsto?sslmode=disable' \
    IMSTO_MAX_FILESIZE=524288 \
    IMSTO_MAX_WIDTH=1920 \
    IMSTO_MAX_HEIGHT=1920 \
    IMSTO_MIN_WIDTH=50 \
    IMSTO_MIN_HEIGHT=50 \
    IMSTO_MAX_QUALITY=88 \
    IMSTO_CACHE_ROOT=/var/lib/imsto/cache/ \
    IMSTO_LOCAL_ROOT=/var/lib/imsto \
    IMSTO_SUPPORT_SIZE="60,120,256" \
    IMSTO_ROOFS="demo" \
    IMSTO_ENGINES="demo:file" \
    IMSTO_STAGE_HOST=""

WORKDIR /opt/imsto
VOLUME ["/var/lib/imsto"]

RUN cat /etc/apk/repositories \
  && sed -ri "s/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/" /etc/apk/repositories \
  && cat /etc/apk/repositories \
  && apk add --update \
    bash \
    postgresql \
    nginx \
    su-exec \
  && mkdir -p /opt/imsto /var/lib/imsto/{stores,cache} /run/nginx \
  && cd /opt/imsto \
  && echo "bundle: imsto bundle" >> Procfile \
  && echo "admin: imsto-admin" >> Procfile \
  && echo "nginx: nginx -g 'daemon off;'" >> Procfile \
  && chown nginx /run/nginx \
  && rm -rf /var/cache/apk/*

ADD database/imsto_*.sql /docker-entrypoint-initdb.d/
COPY --from=build /go/bin/forego /go/bin/imsto /go/bin/imsto-admin /usr/bin/
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-config/host.imsto.docker.conf /etc/nginx/conf.d/default.conf
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-config/entrypoint.sh /ep.sh

EXPOSE 80

ENTRYPOINT ["/ep.sh"]
CMD ["start"]
