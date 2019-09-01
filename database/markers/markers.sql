CREATE EXTENSION postgis;
/*https://stackoverflow.com/questions/24981784/how-do-i-add-postgis-to-postgresql-pgadmin */
CREATE USER docker;
CREATE DATABASE docker;
GRANT ALL PRIVILEGES ON DATABASE docker TO docker;
USE DATABASE docker;
CREATE TABLE markers (
    id          BIGINT PRIMARY KEY, 
    info        TEXT, 
    url         TEXT, 
    title       TEXT, 
    pageid      BIGINT,
    makertype   TEXT, 
    lat         double precision, 
    lon         double precision,
    source      TEXT, 
    beg_year    int, 
    end_year    int
);

SELECT AddGeometryColumn('markers', 'geom', 4326, 'POINT', 2);

