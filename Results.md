| Test  | Pattern type  | Grep args                               | t1 (ms) | t2 (ms) | t3 (ms) | t4 (ms) | t5 (ms) | Avg (ms) | SD (ms) | Notes                 |
|------:|----------------|-----------------------------------------|--------:|--------:|--------:|--------:|--------:|---------:|--------:|-----------------------|
| 1     | Frequent       | -i -e 'POST'                            |      88 |      92 |      94 |      85 |      91 |       90 |    3.16 | 10 workers, 60MB each |
| 2     | Infrequent     | -F -e 'POST /wp-content HTTP/1.0" 200 4964'   |      44 |      44 |      42 |      42 |      43 |       43 |    0.89 | 10 workers, 60MB each |
| 3     | Regex          | -i -E -e '"POST /wp-content HTTP/1\.0"[[:space:]]+200[[:space:]]+4964[[:space:]]+"http://www\.[A-Za-z0-9-]+\.com/index/"' |     131 |     138 |     154 |     128 |     125 |    135.2 |   10.34 | 10 workers, 60MB each |
| 4     | Frequent (1↓)  | -i -e 'POST'                          |   20000 |   20004 |   20000 |   20003 |   20003 |    20002 |    1.67 | One worker killed     |


