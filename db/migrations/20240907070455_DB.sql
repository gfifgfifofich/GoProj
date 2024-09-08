-- migrate:up
CREATE TABLE users
(
    id serial NOT NULL,
    guid text NOT NULL,
    name text NOT NULL,
    password_hash text NOT NULL,
    refreshtokens text[]
);

-- migrate:down

DROP TABLE users;