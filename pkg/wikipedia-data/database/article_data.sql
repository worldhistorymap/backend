CREATE TABLE user_article_data (
  id SERIAL PRIMARY KEY,
  user_id int NOT NULL,
  url TEXT, 
  title TEXT, 
  lat double precision, 
  lon double precision,
  hovered_over int, 
  generated int, 
  clicked int, 
  searched int,
  article_interaction int, 
  created_at TIME,
  updated_at TIME,
  deleted_at TIME
);

CREATE TABLE article_data (
  id SERIAL PRIMARY KEY,
  url TEXT,
  title TEXT,
  lat double precision,
  lon double precision,
  hovered_over int,
  generated int,
  clicked int,
  searched int,
  article_interaction int,
  created_at TIME,
  updated_at TIME,
  deleted_at TIME
);
