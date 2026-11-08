
CREATE TABLE IF NOT EXISTS search_tags.tags
(
    `query` String,
    tag String,
    users UInt64,
    `rank` UInt8
)
ENGINE = MergeTree()
ORDER BY (`query`, `rank`);

TRUNCATE TABLE search_tags.tags;

INSERT INTO search_tags.tags
WITH
normalized AS
(
    SELECT
        user_id,
        timestamp,
        replaceRegexpAll(
            lowerUTF8(trimBoth(search_query)),
            ' +',
            ' '
        ) AS query
    FROM search_tags.search_logs
),

pairs AS
(
    SELECT
        user_id,
        timestamp,
        query,
        leadInFrame(query, 1, '') OVER w AS tag,
        leadInFrame(timestamp, 1) OVER w AS next_time
    FROM normalized
    WINDOW w AS
    (
        PARTITION BY user_id
        ORDER BY timestamp, query
        ROWS BETWEEN UNBOUNDED PRECEDING
                 AND UNBOUNDED FOLLOWING
    )
),

counts AS
(
    SELECT
        query,
        tag,
        uniqExact(user_id) AS users
    FROM pairs
    WHERE query != ''
      AND tag != ''
      AND query != tag
      AND next_time >= timestamp
      AND next_time <= timestamp + INTERVAL 30 MINUTE
    GROUP BY query, tag
    HAVING users >= 5
),

ranked AS
(
    SELECT
        query,
        tag,
        users,
        row_number() OVER
        (
            PARTITION BY query
            ORDER BY users DESC, tag ASC
        ) AS position
    FROM counts
)

SELECT
    query,
    tag,
    users,
    position
FROM ranked
WHERE position <= 10;
