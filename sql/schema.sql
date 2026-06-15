CREATE TABLE IF NOT EXISTS env_master
(
    id
    SERIAL
    PRIMARY
    KEY,
    subdomain
    TEXT
    NOT
    NULL,
    target_url
    TEXT
    NOT
    NULL
);