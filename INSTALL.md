# Installation

## Dependencies
* PostgreSQL 9.4+

### C lib

1. libjpeg-turbo (recommend) or jpeg
2. png (option in plan)

### Go lib

- go version >= 1.11


## Installation

### prepare C libraries
   - osx:
     1. get `libjpeg-turbo-1.4.2.dmg` from http://sourceforge.net/projects/libjpeg-turbo/
     2. mount and install
     or `sudo port install libjpeg-turbo`
   - gentoo:
     - `emerge libjpeg-turbo`
   - debian:
     - `apt-get install libturbojpeg1`
   - alpine:
     - `apk add libjpeg-turbo-dev`

### build

    make


## Launch

### Prepare Database

~~~
psql -Upostgres
-- copy database/schema/pgsql_00_base.sql contents and execute them

-- schema and procedures
cat database/imsto_??_*.sql | psql -Uimsto

forego run ./imsto auth -name demo -save
~~~

### Configuration
- `mkdir /etc/imsto`
- `vim /etc/imsto/imsto.ini`, see also: `demo-config`
- `mkdir /var/log/imsto`

### Launch tiring service
~~~
./imsto tiring
~~~


### Launch stage service
~~~
./imsto stage
~~~

## Change nginx config

- add imsto blocks into locations

~~~
	location ~* ^/(thumb|t2|t3)/(.+\.(?:gif|jpe?g|png))$ {
		alias /opt/imsto/cache/thumb/$2;
		error_page 404 = @imsto_stage;
	}

	location @imsto_stage {
		proxy_pass   http://localhost:8968;
		proxy_set_header   X-Real-IP        $remote_addr;
		proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}

	location /imsto/ {
		proxy_pass   http://localhost:8964;
		proxy_set_header   X-Real-IP        $remote_addr;
		proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}
~~~
