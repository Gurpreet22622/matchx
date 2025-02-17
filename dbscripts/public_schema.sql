set statement_timeout=0;
set lock_timeout=0;
set idle_in_transaction_session_timeout=0;
set client_encoding='UTF-8';
set standard_conforming_strings=on;
set client_min_messages=warning;
set row_security=off;

create extension if not exists plpgsql with schema pg_catalog;
create extension if not exists "uuid-ossp" with schema pg_catalog;

set search_path=public, pg_catalog;
set default_tablespace='';

--users
create table users(
    id uuid not null default uuid_generate_v1mc(),
    username text unique not null,
    full_name text not null,
    user_role text not null,
    email text not null,
    picture text,
    constraint users_pk primary key (id)
);

create index users_username
on users (username);

--property
create table property(
    id uuid not null default uuid_generate_v1mc(),
    user_id uuid not null,
    property_type text not null,
    longitude float not null,
    latitude float not null,
    locality text not null,
    lease_type text not null,
    furnished_status text not null,
    property_area float not null,
    internet boolean not null,
    ac boolean not null,
    ro boolean not null,
    kitchen boolean not null,
    geezer boolean not null,
    constraint property_pk primary key (id),
    constraint fk_user_id foreign key (user_id)
        references users (id) match simple
);

create index users_id
on property (user_id);


--results
-- create table results(
--     id uuid not null default uuid_generate_v1mc(),
--     runner_id uuid not null,
--     race_result interval not null,
--     location text not null,
--     position integer,
--     year integer not null,
--     constraint results_pk primary key (id),
--     constraint fk_results_runner_id foreign key (runner_id)
--         references runners (id) match simple
--         on update no action
--         on delete no action
-- );