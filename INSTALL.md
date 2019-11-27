# Installation

## Dependencies

### Database

* PostgreSQL 9.4+

### Go (>= 1.12)

- go get github.com/ddollar/forego

## Installation

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
- `cp .env.example .env`
- `vim .env`

### Launch tiring service
~~~
foreog run ./imsto tiring
~~~


### Launch stage service
~~~
foreog run ./imsto stage
~~~

## Change nginx config

- add imsto blocks into locations

~~~

	location ~* "^/(show|view)/(\w+)/([a-z0-9]{2})/?([a-z0-9]{2})/?([a-z0-9]{4,36})\.(gif|jpe?g|png|webp)$" {
		alias /opt/imsto/cache/thumb/$2/$3/$4/$5.$6;
		error_page 404 = @imsto_stage;
		expires 1d;
	}

	# frontend for view
	location @imsto_stage {
		proxy_pass   http://localhost:8968;
		proxy_set_header   X-Real-IP        $remote_addr;
		proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}

	# backend for management
	location /imsto/ {
		proxy_pass   http://localhost:8964;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection $http_connection;
        proxy_set_header   X-Scheme $scheme;
        proxy_set_header   X-Real-IP        $remote_addr;
        proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}

	location / {
		proxy_pass   http://localhost:8970;
		proxy_set_header   X-Real-IP        $remote_addr;
		proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}

~~~
