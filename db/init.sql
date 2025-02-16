create database shop;

\c shop

create table if not exists users (
    id integer primary key generated always as identity,
    name text check(length(name) >= 3 and length(name) < 150) unique not null,
    password text not null,
    money integer check(money >= 0) default 1000,
    registered_at timestamp default now() not null
);

create index users_name on users using gin(to_tsvector('english', name));

create table if not exists product (
    id integer primary key generated always as identity,
    name text not null,
    price integer check(price >= 1) default 1 not null
);

insert into product(name, price) values
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

create table if not exists user_transaction (
    user_from integer,
    user_to integer,
    money integer not null,
    sent_at timestamp default now() not null,
    foreign key (user_from) references users(id) on delete set null,
    foreign key (user_to) references users(id) on delete set null
);

create index user_transaction_user_from_time on user_transaction(user_from, sent_at);
create index user_transaction_user_to_time on user_transaction(user_to, sent_at);

create table if not exists user_product (
    user_id integer,
    product_id integer,
    bought_at timestamp default now() not null,
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (product_id) references product(id) on delete set null
);

create index user_product_user_time on user_product(user_id, bought_at);
