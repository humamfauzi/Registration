package Registration

import (
  "net/http"
  "net/http/httptest"

  "testing"
  "bytes"
  "io/ioutil"
  "strings"

  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

type routeTest struct {
  // From test in gorilla/mux

  title       string
  body        string
  from        string
  method      string
  handler     http.Handler
  shouldMatch bool
}

// +-------------------------------------------------+
//                 TESTING SEQUENCE
// +-------------------------------------------------+

func TestMain(t *testing.T) {
  //Implement

  err := SummonDatabase()
  if err != nil {
    t.Errorf("Problem Creating Database %v", err)
    return
  }

  aes := GetValue("./jsonFiles/config.json", "AES")
  conn0, err := sql.Open("mysql", GetValue("./jsonFiles/config.json", "access0"))
  if err != nil {
    t.Errorf("DATABASE CONNECTION FAILURE. %v", err)
    return
  }

  tests := []routeTest{
    {
      title: "Register--00",
      body: `{"name": "Hasan0", "email": "contact1@hasan.com", "phone": "+83774322321", "password": "Hasan123"}`,
      from: "http://example.com/foo",
      method: "POST",
      handler: CreateNewUser{db: conn0, aesCredentials: aes},
      shouldMatch: true,
    },
    {
      title: "Register--01",
      body: `{"name": "Jailani0", "email": "contact1@jailani.com", "phone": "+9312351342", "password":"ae3ne5fds"}`,
      from: "http://example.com/foo",
      method: "POST",
      handler: CreateNewUser{db: conn0, aesCredentials: aes},
      shouldMatch: true,
    },
    {
      title: "Register--02",
      body: `{"name": "Charles0","email": "contact1@charles.com", "password":"Bhs^23u8}[]"}`,
      from: "http://example.com/foo",
      method: "POST",
      handler: CreateNewUser{db: conn0, aesCredentials: aes},
      shouldMatch: false,
    },
  }

  for _, test := range tests {
    testRoute(t, test)
  }

  conn0.Close()
  return
}

func VerifyUserTest(t *testing.T) {
  // Implement
  return
}

func LoginTest(t *testing.T) {
  // Implement
  return
}

func ForgetPassTest(t *testing.T) {
  // Implement
  return
}

func VerifyTokenTest(t *testing.T) {
  // Implement
  return
}

func PassRecoveryTest(t *testing.T) {
  // Implement
  return
}

// +--------------------------------------------------------+
//                     FUNCTION HELPER
// +--------------------------------------------------------+
func SummonDatabase() error {
  conn0, err := sql.Open("mysql","root:humam123@tcp(localhost:3306)/")
  if err != nil {
    return err
  }

  _, err = conn0.Exec("CREATE DATABASE IF NOT EXISTS registration;")
  if err != nil {
    return err
  }

  conn0.Close()

  conn0, err = sql.Open("mysql", GetValue("./jsonFiles/config.json", "access0"))
  if err != nil {
    return err
  }

  query, err := ioutil.ReadFile("./db/migrations/tableCreation.sql")
  if err != nil {
    return err
  }

  splitQuery := strings.Split(string(query), "\n\n")
  for _, v := range splitQuery {
    _, err = conn0.Exec(v)
    if err != nil {
      return err
    }
  }

  conn0.Close()
  return nil
}

func testRoute(t *testing.T, test routeTest) {
  method := test.method
  origin := test.from
  body := []byte(test.body)

  req := httptest.NewRequest(method, origin, bytes.NewReader(body))
  w := httptest.NewRecorder()

  reg := test.handler
  reg.ServeHTTP(w, req)

  response := w.Result()
  isSuccess := response.StatusCode == 200

  if isSuccess != test.shouldMatch {
    t.Errorf("-- %v -- return false answer: expected %v, got %v", test.title, test.shouldMatch, isSuccess)
  }
}
