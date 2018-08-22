-- DDL generated by Postico 1.4.2
-- Not all database features are supported. Do not use for backup.

-- Table Definition ----------------------------------------------

CREATE TABLE if not exists provision (
    name character varying(200) PRIMARY KEY,
    plan character varying(200),
    claimed character varying(200),
    make_date timestamp without time zone DEFAULT now()
);

-- Indices -------------------------------------------------------

CREATE UNIQUE INDEX if not exists name_pkey ON provision(name text_ops);
