create table "subscription"
(
    id           serial primary key,
    service_name varchar(1000)      not null,
    price        integer            not null,
    user_id      uuid               not null,
    start_date   date default now() not null,
    end_date     date,
    delete_date  date
);

create index idx_subs_user_id
    on subscription (user_id);

create unique index idx_subs_service_name_user_id_uindex
    on subscription (service_name, user_id);