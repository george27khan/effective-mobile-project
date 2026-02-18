create table "subscription"
(
    id           serial primary key,
    service_name varchar(1000)      not null,
    price        integer            not null,
    user_id      uuid               not null,
    start_date   date default now() not null,
    end_date     date,
    is_delete    boolean default false not null
);

create unique index idx_subs_service_name_user_id_uindex
    on subscription (user_id, service_name);

---- create above / drop below ----

drop table "subscription";