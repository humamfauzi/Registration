package main

import (
  "encoding/json"
  "io/ioutil"
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
  var dict Dictionary
  json.Unmarshal(body, &dict)

  host := dict[emailType].(map[string]interface{})["host"].(string)
  addr := dict[emailType].(map[string]interface{})["addr"].(string)
  pass := dict[emailType].(map[string]interface{})["pass"].(string)
  port := dict[emailType].(map[string]interface{})["port"].(string)

  return host, addr, pass, port
}

func DatabaseInsert(db *sql.DB, query string, input ...interface{}) error {
  stmt, err = db.Prepare(query)
  if err != nil {
    log.Fatal(err)
    return err
  }

  rows, err = stmt.Exec(input...)
  if err != nil {
    log.Fatal(err)
    return err
  }

  _, err = rows.LastInsertId()
  if err != nil {
    log.Fatal(err)
    return err
  }
}
