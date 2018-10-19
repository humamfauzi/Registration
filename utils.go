package main

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"

  "encoding/json"
  "io/ioutil"
  "log"
)

type Dictionary map[string]string

func GetValue(directory string, key string) string {
  body, _ := ioutil.ReadFile(directory)
  var dict Dictionary
  json.Unmarshal(body, &dict)
  return dict[key]
}

func GetValueEmail(directory, emailType string) (string, string, string, string) {
  body, _ := ioutil.ReadFile(directory)

  var dict map[string]interface{}
  json.Unmarshal(body, &dict)

  buffer0 := dict[emailType].(map[string]interface{})

  host := buffer0["host"].(string)
  addr := buffer0["addr"].(string)
  pass := buffer0["pass"].(string)
  port := buffer0["port"].(string)

  return host, addr, pass, port
}

func DatabaseInsert(db *sql.DB, query string, input ...interface{}) error {
  stmt, err := db.Prepare(query)
  if err != nil {
    log.Fatal(err)
    return err
  }

  rows, err := stmt.Exec(input...)
  if err != nil {
    log.Fatal(err)
    return err
  }

  _, err = rows.LastInsertId()
  if err != nil {
    log.Fatal(err)
    return err
  }
  return nil
}

func ReadQuery(db *sql.DB, query string, input ...interface{}) ([]interface{}, error) {
  rows, _ := db.Query(query, input...)
  columns, _ := rows.Columns()
  count := len(columns)

  val := make([]interface{}, count)
  valPtrs := make([]interface{}, count)

  for rows.Next() {
    for i, _ := range columns {
      valPtrs[i] = &val[i]
    }
    rows.Scan(valPtrs...)
    return val, nil
  }
  return nil, nil
}
