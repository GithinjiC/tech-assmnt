-- # Steps:
-- # 1. Run EXPLAIN and EXPLAIN ANALYZE to figure out the query's execution plan. 
-- # 2. Verify indexes and joins

-- # 3. Analyze the performance of the SQL statement
-- This query will identify 20 queries with the highest execution time and provide impact on system perfomance
SELECT substring(query, 1, 80) AS short_query,
              round(total_exec_time::numeric, 2) AS total_exec_time,
              calls,
              round(mean_exec_time::numeric, 2) AS mean,
              round((100 * total_exec_time /
              sum(total_exec_time::numeric) OVER ())::numeric, 2) AS percentage_cpu
FROM    pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 30;

-- # 4. This query returns long running queries i.e more than 10 minutes
SELECT pid, usename, state, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '10 minutes';

-- # 5. This query returns queries that are blocked by other queries. Look out for and analyze the locks
SELECT pid, usename, client_addr, client_hostname, application_name, pg_blocking_pids(pid) as blocked_by, query as blocked_query FROM pg_stat_activity WHERE cardinality(pg_blocking_pids(pid)) > 0;

-- # 6. Consider the need for vacuuming/autovacuuming

-- # 7. Track cache efficiency. The higher the cache miss ration, the higher the disk reads.