
CREATE DATABASE IF NOT EXISTS search_tags;

CREATE TABLE IF NOT EXISTS search_tags.search_logs
(
    timestamp DateTime64(3),
    user_id UInt64,
    search_query String
)
ENGINE = MergeTree()
ORDER BY (user_id, timestamp);

INSERT INTO search_tags.search_logs
SELECT
    assumeNotNull(timestamp),
    assumeNotNull(user_id),
    assumeNotNull(search_query)
FROM file('logs.parquet', 'Parquet');
