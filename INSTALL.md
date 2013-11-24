# Installation

## Dependencies

### C lib

1. jpeg
2. png (option)
3. ImageMagick

### Go lib

- go version >= 1.2

~~~
go get github.com/vaughan0/go-ini
go get github.com/lib/pq
go get github.com/mitchellh/mapstructure
go get github.com/crowdmob/goamz/s3
~~~


## Installation

### prepare C libraries
   - osx:
     - `port install jpeg`
     - `port install ImageMagick`
   - gentoo:
     - `emerge jpeg`
     - `emerge imagemagick`

### get and build

    go get wpst.me/calf/imsto


## Launch

### Launch tiring backend
~~~
IMSTO_API_0_SALT=$SALT $GOPATH/bin/imsto -root=$APP_ROOT -logs=$LOG_ROOT tiring -port 5564
~~~


### Launch stage backend
~~~
$GOPATH/bin/imsto -root=$APP_ROOT -logs=$LOG_ROOT stage -port 5580
~~~

## Change nginx config

- add imsto blocks into locations

~~~
	location /thumb {
		alias /opt/imsto/cache/thumb/;
		error_page 404 = @imsto_stage;
	}

	location @imsto_stage {
		proxy_pass   http://localhost:5580;
		proxy_set_header   X-Real-IP        $remote_addr;
		proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}

	location /imsto {
		proxy_pass   http://localhost:5564;
		proxy_set_header   X-Real-IP        $remote_addr;
		proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
	}
~~~
