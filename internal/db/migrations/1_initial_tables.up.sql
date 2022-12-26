create table if not exists users (
    id int auto_increment primary key,
    discord_id varchar(255) UNIQUE,
    email varchar(255) UNIQUE
);
