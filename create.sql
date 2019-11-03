DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS posts CASCADE;

CREATE TABLE users (
                     id serial PRIMARY KEY,
                     email       varchar not null CHECK (email <> ''),
                     password    varchar not null CHECK (password <> ''),
                     created_at timestamp with time zone NOT NULL default now(),
                     updated_at timestamp with time zone NOT NULL default now(),
                     deleted_at timestamp with time zone default null
);

CREATE TABLE posts (
                     id serial PRIMARY KEY,
                     user_id integer references users(id),
                     title varchar check (title <> ''),
  -- title or body can be absent, but not both. if either is present, it can't be empty
                     body varchar check (body <> '' and ((body is not null) or (title is not null))),
                     created_at timestamp with time zone NOT NULL default now(),
                     updated_at timestamp with time zone NOT NULL default now(),
                     deleted_at timestamp with time zone default null
);
