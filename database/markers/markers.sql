/**We use a prebuilt docker image with postgis already installed **/
GRANT ALL PRIVILEGES ON DATABASE gis TO docker;
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

