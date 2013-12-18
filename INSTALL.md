# Installation

## Dependencies
* PostgreSQL 9+
* PostgreSQL extension: `hstore`

### C lib

1. libjpeg-turbo (recommend) or jpeg
2. png (option in plan)
3. ~~ImageMagick~~ (*not need anymore*)

### Go lib

- go version >= 1.2

~~~
go get github.com/vaughan0/go-ini
go get github.com/lib/pq
go get github.com/mitchellh/mapstructure
go get github.com/crowdmob/goamz/s3
go get github.com/nfnt/resize
~~~


## Installation

### prepare C libraries
   - osx:
     1. get `libjpeg-turbo` from http://sourceforge.net/projects/libjpeg-turbo/
     2. extract and build:
     3. `./configure --prefix=/usr --with-jpeg8 --host x86_64-apple-darwin && make && sudo make install`
   - gentoo:
     - `emerge libjpeg-turbo`

### get and build

    go get wpst.me/calf/imsto


## Launch

### Prepare Database

~~~
psql -Upostgres
-- copy database/schema/pgsql_00_base.sql contents and execute them

-- basic schema
psql -Uimsto -f database/schema/pgsql_10_imsto.sql

-- mobile upload schema
psql -Uimsto -f database/schema/pgsql_11_imsto_auth.sql

-- demo schema
psql -Uimsto -f database/schema/pgsql_12_imsto_demo.sql

-- procedure
psql -Uimsto database/procedure/pgsql_10_imsto.sql
psql -Uimsto database/procedure/pgsql_11_imsto_auth.sql
~~~

### Configuration
- `mkdir /etc/imsto`
- `vim /etc/imsto/imsto.ini`, see also: `demo-config`
- `mkdir /var/log/imsto`

### Launch tiring service
~~~
IMSTO_API_0_SALT=mysalt $GOPATH/bin/imsto tiring
~~~


### Launch stage service
~~~
$GOPATH/bin/imsto stage
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
