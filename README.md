# Imsto redesigned in Golang


## Launch with docker

```sh

# postgresql
docker create --name imsto-db-data -v /var/lib/postgresql busybox:1 echo imsto db data
docker run --name imsto-db -p 54323:5432 \
	-e DB_PASS=mypassword \
	-e TZ=Hongkong \
	--volumes-from=imsto-db-data \
	-d liut7/imsto-db:latest

# main conatainers
docker run --name imsto --rm -p 8180:80 \
	--link imsto-db -v /private/var/lib/imsto:/var/lib/imsto\
	-d liut7/imsto

```

## Installation

> see more: INSTALL.md

