| Test  | Pattern type  | Grep args                                | t1 (ms) | t2 (ms) | t3 (ms) | t4 (ms) | t5 (ms) | Avg (ms) | SD (ms) | Notes                    |
|------:|----------------|-------------------------------------------|---------:|---------:|---------:|---------:|---------:|---------:|--------:|-------------------------|
| 1     | Frequent       | -i -e 'error'                             |         |         |         |         |         |         |        | 4 workers, 60MB each    |
| 2     | Infrequent     | -F -e 'very_unlikely_substring_12345'     |         |         |         |         |         |         |        | 4 workers, 60MB each    |
| 3     | Regex          | -i -E -e 'POST /wp-content HTTP/1\\.0…'   |         |         |         |         |         |         |        | 4 workers, 60MB each    |
| 4     | Frequent (1↓)  | -i -e 'error'                             |         |         |         |         |         |         |        | One worker killed       |


