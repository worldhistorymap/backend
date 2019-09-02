FROM kartoza/postgis:11.0-2.5
COPY markers.sql /docker-entrypoint-initdb.d/
VOLUME /var/lib/postgresql/data
EXPOSE 5432
