CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO subscription (service_name, price, user_id, start_date, end_date, is_delete)
SELECT
    ('Plan_' || (trunc(random()*5)+1))::varchar,

    -- разные диапазоны цен
    CASE
        WHEN random() < 0.25 THEN (trunc(random()*500)+100)::int
        WHEN random() < 0.50 THEN (trunc(random()*2000)+1000)::int
        WHEN random() < 0.75 THEN (trunc(random()*5000)+3000)::int
        ELSE (trunc(random()*10000)+8000)::int
        END AS price,

    u.user_id,

    -- 80% раньше текущего месяца, 20% текущий месяц
    CASE
        WHEN random() < 0.80 THEN
            (
                date_trunc('month', current_date)
                    - (trunc(random()*12 + 1) || ' month')::interval
                )::date
        ELSE
            date_trunc('month', current_date)::date
        END AS start_date,

    -- 30% имеют end_date
    CASE
        WHEN random() < 0.30 THEN
            (
                        current_date
                    + (CASE WHEN random() < 0.5
                                THEN trunc(random()*60)
                            ELSE -trunc(random()*60)
                           END || ' day')::interval
                )::date
        ELSE NULL
        END AS end_date,

    -- 10% удалённых
    (random() < 0.10) AS is_delete

FROM (
         SELECT gen_random_uuid() AS user_id
         FROM generate_series(1,50)
     ) u
         CROSS JOIN LATERAL
    generate_series(1, (trunc(random()*10)+1)::int)

ON CONFLICT (user_id, service_name) DO NOTHING;

---- create above / drop below ----

truncate table "subscription"


