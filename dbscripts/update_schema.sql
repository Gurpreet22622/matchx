-- create extension if not exists pgcrypto;

-- create table users(
--     id uuid not null default uuid_generate_v1mc(),
--     username text not null unique,
--     user_password text not null,
--     user_role text not null,
--     access_token text,
--     constraint users_pk primary key (id) 
-- );

-- create index user_access_token
-- on users (access_token);

-- insert into users(username, user_password,user_role)
-- values
-- ('admin',crypt('admin',gen_salt('bf')),'admin'),
-- ('runner',crypt('runner',gen_salt('bf')),'runner');
