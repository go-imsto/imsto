#!/bin/bash
set -e


[ -z "$PGDATA" ] && echo "PGDATA not set" && exit 1
[ -d "$PGDATA" ] || mkdir -p "$PGDATA"
chown -R postgres:postgres "$PGDATA"

mkdir -p /run/postgresql
chmod g+s /run/postgresql
chown -R postgres:postgres /run/postgresql

if [ -z "$(ls -A "$PGDATA")" ]; then
	# su-exec postgres initdb
	echo "start initdb"
	su-exec postgres:postgres initdb -D "${PGDATA}" $PG_INITDB_OPTS

	sed -ri "s/^#(listen_addresses\s*=\s*)\S+/\1'*'/" "$PGDATA"/postgresql.conf

	if [ -n "$PG_AUTOVACUUM" ]; then
		sed -ri "s/^#(autovacuum\s*=\s*)\S+/\1on/" "$PGDATA"/postgresql.conf
	fi

	# check password first so we can ouptut the warning before postgres
	# messes it up
	if [ -n "$DB_PASS" ]; then
		pass="PASSWORD '$DB_PASS'"
		authMethod=md5
	else
		# The - option suppresses leading tabs but *not* spaces. :)
		cat >&2 <<-'EOWARN'
			****************************************************
			WARNING: No password has been set for the database.
			         This will allow anyone with access to the
			         Postgres port to access your database. In
			         Docker's default configuration, this is
			         effectively any other container on the same
			         system.

			         Use "-e DB_PASS=password" to set
			         it in "docker run".
			****************************************************
		EOWARN

		pass=
		authMethod=trust
	fi

	{ echo; echo "host all all 0.0.0.0/0 $authMethod"; } >> "$PGDATA"/pg_hba.conf

	# internal start of server in order to allow set-up using psql-client
	# does not listen on external TCP/IP and waits until start finishes
	su-exec postgres:postgres pg_ctl -D "$PGDATA" \
		-o "-c listen_addresses='localhost'" \
		-w start

	: ${DB_USER:=postgres}
	: ${DB_NAME:=$DB_USER}

	psql=( psql -v ON_ERROR_STOP=1 )

	if [ -n "$DB_USER" -a "$DB_USER" != 'postgres' ]; then
		echo "Creating user \"${DB_USER}\"..."
		echo "CREATE ROLE ${DB_USER} with LOGIN CREATEDB $pass ;" |
			"${psql[@]}" -U postgres >/dev/null

	fi

	if [ -n "$DB_NAME" -a "$DB_NAME" != 'postgres' ]; then
		echo "Creating database \"${DB_NAME}\"..."
		echo "CREATE DATABASE $DB_NAME WITH OWNER = ${DB_USER} ENCODING = 'UTF8' ;" |
			"${psql[@]}" -U postgres >/dev/null

		echo "Granting access to database \"${DB_NAME}\" for user \"${DB_USER}\"..."
		echo "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME to ${DB_USER};" |
			"${psql[@]}" -U postgres >/dev/null

		if [ -n "$DB_SCHEMA" -a "$DB_SCHEMA" != 'public' ]; then
			echo "CREATE SCHEMA ${DB_SCHEMA} for DB ${DB_NAME}"
			echo "CREATE SCHEMA ${DB_SCHEMA} AUTHORIZATION ${DB_USER};" |
				"${psql[@]}" -U postgres ${DB_NAME} >/dev/null
		fi
	fi

	psql+=( -U "$DB_USER" -d "$DB_NAME" )

	for f in /docker-entrypoint-initdb.d/*; do
		case "$f" in
			*.sh)     echo "$0: running $f"; . "$f" ;;
			*.sql)    echo "$0: running $f"; "${psql[@]}" < "$f" && echo ;;
			*.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
			*)        echo "$0: ignoring $f" ;;
		esac
		echo
	done

	su-exec postgres pg_ctl -D "$PGDATA" -m fast -w -t 2 stop
fi

if [ "$1" = 'start' ]; then
	su-exec postgres:postgres pg_ctl -D "$PGDATA" \
		-o "-c listen_addresses='${PG_LISTEN:-localhost}'" \
		-w -t 2 start

	sleep 2

	if [[ ! -s htdocs/api_key.js ]]; then
		echo "add demo api_key"
		echo var api_key=\"$(imsto auth -name demo -save | grep 'key:' | awk '{print $2}')\"\; > htdocs/api_key.js
	fi

	chown -R nobody /var/lib/imsto
	forego start
else
	exec "$@"
fi

