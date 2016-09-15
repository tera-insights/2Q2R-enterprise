#/bin/bash -e

touch test.db

sqlite3 test.db "CREATE TABLE app_infos(id INTERGER,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  app_id VARCHAR,
  app_name VARCHAR);"
sqlite3 test.db "INSERT INTO app_infos VALUES(4, '',
  '', '','_T-zi0wzr7GCi4vsfsXsUuKOfmiWLiHBVbmJJPidvhA','test-app');"
