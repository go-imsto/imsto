FROM lcgc/postgresql:9.6.13

ENV DB_NAME=imsto DB_USER=imsto

ADD imsto_*.sql /docker-entrypoint-initdb.d/
