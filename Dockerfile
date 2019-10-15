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
  && go mod download \
  && export LDFLAGS="-X ${ROOF}/cmd.VERSION=$(date '+%Y%m%d')" \
  && env \
  && GOOS=linux go install -ldflags "${LDFLAGS} -s -w" . \
  && echo "build done"


FROM alpine:3.6

ENV PGHOST="imsto-db" \
    IMSTO_META_DSN='postgres://imsto:mypassword@imsto-db/imsto?sslmode=disable' \
    IMSTO_ROOT="/opt/imsto"

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
  && mkdir -p /opt/imsto/config /var/lib/imsto /run/nginx \
  && cd /opt/imsto \
  && echo "stage: imsto -conf /opt/imsto/config stage" >> Procfile \
  && echo "tiring: imsto -conf /opt/imsto/config tiring" >> Procfile \
  && echo "nginx: nginx -g 'daemon off;'" >> Procfile \
  && chown nginx /run/nginx \
  && rm -rf /var/cache/apk/*

COPY --from=build /go/bin/forego /go/bin/imsto /usr/bin/
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-config/imsto.ini /opt/imsto/config/
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-config/host.imsto.docker.conf /etc/nginx/conf.d/default.conf
COPY --from=build /go/src/github.com/go-imsto/imsto/apps/demo-site /opt/imsto/htdocs

EXPOSE 80

CMD ["forego", "start"]
