package main

import (
  "net/http"
  "net/url"
  "net/smtp"
  "io/ioutil"

  "encoding/json"
  "encoding/hex"
  "crypto/sha512"

  "time"
  "strconv"
  "strings"
  "log"

  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

type Registration struct {
  Name     string `json: name`
  Email    string `json: email`
  Phone    string `json: phone`
  Password string `json: password`
}

type Login struct {
  Email string `json: email`
  Password string `json: password`
}

type Cookie struct {
  Name       string
  Value      string
  Path       string
  Domain     string
  Expires    time.Time
  RawExpires string
  MaxAge     int
  Secure     bool
  Httponly   bool
  Raw        string
  Unparsed   []string
}

type CreateNewUser struct {
  db *sql.DB
  aesCredentials string
}

func (cnu *CreateNewUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  var regis Registration

  err = json.Unmarshal(body, &regis)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusBadRequest)
  }

  query := GetValue("./query.json", "CreateNewUser")
  err = DatabaseInsert(cnu.db, query, regis.Name, regis.Phone, regis.Email)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusBadRequest)
  }

  query = GetValue("./query.json", "CreateNewPassword")
  err = DatabaseInsert(cnu.db, query, regis.Email, regis.Email + ":" + regis.Password, cnu.aesCredentials)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusBadRequest)
  }
  w.WriteHeader(http.StatusOK)
}

type LoginUser struct {
  db *sql.DB
  aesCredentials string
}

func (lu *LoginUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  var login Login

  err = json.Unmarshal(body, &login)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  expiration := time.Now().add(30 * 24 * time.Hour)
  cookieToken := lu.CookieValue(expiration, login.Email)
  cookieValue := hex.EncodeToString(cookieValue)

  cookie := Cookie{Name: login.Email, Value: cookieValue, Expires: expiration}
  err = lu.SendCookieToDB(cookie)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
  }
  http.setCookie(w, &cookie)

  query := GetValue("./query.json", "LoginCredentials")
  row := lu.db.QueryRow(query, login.Email + ":" + login.Password, lu.aesCredentials)

  var pswdBlob []byte

  switch err := row.Scan(&pswdBlob); err {
  case sql.ErrNoRows:
    w.WriteHeader(http.StatusNotFound)
  case nil:
    w.WriteHeader(http.StatusOK)
  default:
    w.WriteHeader(http.StatusNotFound)
  }
  w.WriteHeader(http.StatusNotFound)
}

func (lu *LoginUser) CookieValue(email string, expiration time.Time) []byte {
  byteVersion := string(expiration.UnixNano())

  newToken := []byte(byteVersion + email)

  hash := sha256.New()
  hash.Write(newToken)

  return hash.Sum(nil)
}

func (lu *LoginUser) SendCookieToDB(cookie Cookie) error {
  query := GetValue("./query.json", "RegisterCookie")
  err := DatabaseInsert(lu.db, query, cookie.Expires, cookie.Value, cookie.Name)
  if err != nil {
    log.Fatal(err)
    return err
  }
  return nil
}

type ForgetPass struct {
  db *sql.DB
}

func (fp *ForgetPass) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  var email string

  err = json.Unmarshal(body, &email)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  query := GetValue("./query.json", "ForgetEmail")
  row := fp.db.QueryRow(query, email)

  var name string

  switch err := row.Scan(&email, &name); err {
  case sql.ErrNoRows:
    w.WriteHeader(http.StatusNotFound)
  case nil:

    err := fp.CreateToken(email, name)
    if err != nil {
      log.Fatal(err)
      w.WriteHeader(http.StatusNotFound)
    }

  default:
    w.WriteHeader(http.StatusNotFound)
  }
  w.WriteHeader(http.StatusNotFound)
}

func (fp *ForgetPass) CreateToken(email, name string) error {
  // You can change your Token mechanics here
  timeNow := time.Now().UnixNano()
  timeString := strconv.Itoa(int(timeNow))

  hash := sha512.New()
  hash.Write([]byte(email + ":" + timeString))

  token := hash.Sum(nil)

  // Send with Email
  err := fp.SendEmail(email, name, token)
  if err != nil {
    return err
  }

  // Send to Database
  query := GetValue("./query.json", "CreateTokenForgetPass")
  err = DatabaseInsert(fp.db, query, email, timeString, token)
  if err != nil {
    return err
  }
  return nil
}

func (fp *ForgetPass) SendEmail(email, name string, token []byte) error {
  host, addr, pass, port := GetValueEmail("./config.json", "PassRecovery")
  auth := smtp.PlainAuth("", addr, pass, host)

  // Email Link should contain the email and token for later parsing
  template := GetValue("./template.json", "ForgetPass")

  link, _ := url.Parse("https://localhost:8080/VerifyToken?")
  q := link.Query()
  q.Set("Token", hex.EncodeToString(token)) // Change First from HEX to String
  link.RawQuery = q.Encode()

  msg := []byte(fp.ComposeMessage(template, name, link.String()))

  err := smtp.SendMail(host + ":" + port, auth, addr, []string{email}, msg)
  if err != nil {
    log.Fatal(err)
    return err
  }
  return nil
}

func (fp *ForgetPass) ComposeMessage(template, name, link string) string {
  r := strings.NewReplacer("${LINK}", link, "${NAME}", name)
  return r.Replace(template)
}

type VerifyToken struct {
  db *sql.DB
}

func (vt *VerifyToken) ServeHTTP(w http.ResponseWriter, r *http.Request) {

  queryString, err := url.ParseQuery(r.URL.RawQuery)
  if err != nil {
    log.Fatal("FAILED TO PARSE QUERY STRING:", err)
    w.WriteHeader(http.StatusBadRequest)
  }

  query := GetValue("./query.json", "VerifyEmail")
  row := vt.db.QueryRow(query, queryString["token"])

  var token []byte
  var email string

  switch err := row.Scan(&token, &email); err {
  case sql.ErrNoRows:
    w.WriteHeader(http.StatusNotFound)
  case nil:
    w.WriteHeader(http.StatusOK)
  default:
    w.WriteHeader(http.StatusNotFound)
  }
  w.WriteHeader(http.StatusNotFound)
}

type PasswordRecovery struct {
  db *sql.DB
  aesCredentials string
}

func (pr *PasswordRecovery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if r.Header["Content-Type"][0] != "application/json" {
    w.WriteHeader(http.StatusBadRequest)
  }

  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
  }

  var profile Login

  err = json.Unmarshal(body, &profile)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
  }

  query := GetValue("./query.json", "UpdatePassword")
  err = DatabaseInsert(pr.db, query, profile.Email + ":" + profile.Password, pr.aesCredentials)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  query = GetValue("./query.json", "DeleteToken")
  err = DatabaseInsert(pr.db, query, profile.Email + ":" + profile.Password, pr.aesCredentials)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  w.WriteHeader(http.StatusOK)
}
