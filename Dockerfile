FROM golang:1.13-alpine as build
MAINTAINER Eagle Liut <eagle@dantin.me>

RUN cat /etc/apk/repositories \
  && sed -ri "s/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/" /etc/apk/repositories \
  && cat /etc/apk/repositories \
  && apk add --update \
  build-base \
  libjpeg-turbo-dev \
  && rm -rf /var/cache/apk/*

ENV GO111MODULE=on GOPROXY=https://goproxy.io ROOF=github.com/go-imsto/imsto
WORKDIR /go/src/$ROOF
ADD . /go/src/$ROOF

RUN go version \
  && go get github.com/ddollar/forego \
  && export LDFLAGS="-X ${ROOF}/cmd.VERSION=$(date '+%Y%m%d')" \
  && env \
  && GOOS=linux go install -ldflags "${LDFLAGS} -s -w" . \
  && echo "build done"


FROM alpine:3.10

ENV PGHOST="imsto-db" \
    IMSTO_META_DSN='postgres://imsto:mypassword@imsto-db/imsto?sslmode=disable' \
    IMSTO_MAX_FILESIZE=524288 \
    IMSTO_MAX_WIDTH=1920 \
    IMSTO_MAX_HEIGHT=1920 \
    IMSTO_MIN_WIDTH=50 \
    IMSTO_MIN_HEIGHT=50 \
    IMSTO_MAX_QUALITY=88 \
    IMSTO_CACHE_ROOT=/var/lib/imsto/cache/ \
    IMSTO_TEMP_ROOT=/var/lib/imsto/tmp/ \
    IMSTO_SUPPORT_SIZE="60,120,256" \
    IMSTO_SECTIONS="demo:LocalDemo" \
    IMSTO_ENGINES="demo:file" \
    IMSTO_STAGE_HOST="man.imsto.net" \
    IMSTO_LOCAL_ROOT=/var/lib/imsto/stores

WORKDIR /opt/imsto
VOLUME ["/var/lib/imsto"]

RUN cat /etc/apk/repositories \
  && sed -ri "s/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/" /etc/apk/repositories \
  && cat /etc/apk/repositories \
  && apk add --update \
    bash \
    libjpeg-turbo-dev \
    nginx \
    su-exec \
  && mkdir -p /opt/imsto /var/lib/imsto/{stores,cache,tmp} /run/nginx \
  && cd /opt/imsto \
  && echo "rpc: imsto rpc" >> Procfile \
  && echo "stage: imsto stage" >> Procfile \
  && echo "tiring: imsto tiring" >> Procfile \
  && echo "nginx: nginx -g 'daemon off;'" >> Procfile \
  && chown nginx /run/nginx \
  && rm -rf /var/cache/apk/*

COPY --from=build /go/bin/forego /go/bin/imsto /usr/bin/
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-config/host.imsto.docker.conf /etc/nginx/conf.d/default.conf
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-site /opt/imsto/htdocs

EXPOSE 80

CMD ["forego", "start"]
