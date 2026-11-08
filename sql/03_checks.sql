
SELECT count() AS same_query_tag
FROM search_tags.tags
WHERE query = tag;

SELECT count() AS users_below_five
FROM search_tags.tags
WHERE users < 5;

SELECT count() AS queries_over_ten_tags
FROM
(
    SELECT query, count() AS tag_count
    FROM search_tags.tags
    GROUP BY query
    HAVING tag_count > 10
);
