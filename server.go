package main

import (
  "log"
  "net/http"
  "github.com/gorilla/mux"

  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

/*
  Basically what we want to achieve is basic internet Login Mechanism
  which widely applied throughout internet industries.
*/
func main() {
  router, err := NewRouter()
  if err != nil {
    log.Fatal("FAILED TO START SERVER. ", err)
  }

  log.Printf("Starting Server...")
  log.Fatal(http.ListenAndServe(":8080", router))
}

func NewRouter() (*mux.Router, error) {
  muxRouter := mux.NewRouter().StrictSlash(true)

  conn0, err := sql.Open("mysql", GetValue("./config.json", "access0"))
  if err != nil {
    log.Fatal("DATABASE CONNECTION FAILURE. ", err)
    return muxRouter, err
  }

  aes := GetValue("./config.json", "AES")

  muxRouter.Handle("/Register", &CreateNewUser{db: conn0, aesCredentials: aes}).Methods("POST")
  muxRouter.Handle("/Login", &LoginUser{db: conn0, aesCredentials: aes}).Methods("POST")
  muxRouter.Handle("/ForgetPass", &ForgetPass{db: conn0}).Methods("POST")
  muxRouter.Handle("/VerifyToken", &VerifyToken{db: conn0}).Methods("GET")
  muxRouter.Handle("/Passrecovery", &PasswordRecovery{db: conn0, aesCredentials: aes}).Methods("POST")
  return muxRouter, nil
}
