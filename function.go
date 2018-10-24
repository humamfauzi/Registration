package Registration

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
  Email    string `json: email`
  Password string `json: password`
}

type CreateNewUser struct {
  db             *sql.DB
  aesCredentials string
}

func (cnu CreateNewUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
    return
  }

  var regis Registration

  err = json.Unmarshal(body, &regis)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  query := GetValue("./jsonFiles/query.json", "CreateNewUser")
  err = DatabaseInsert(cnu.db, query, regis.Name, regis.Phone, regis.Email)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  query = GetValue("./jsonFiles/query.json", "CreateNewPassword")
  err = DatabaseInsert(cnu.db, query, regis.Email, regis.Email + ":" + regis.Password, cnu.aesCredentials)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  w.WriteHeader(http.StatusOK)
  return
  // if err := SendEmail(regis.Email, regis.Name, ); err != nil {
  //
  // }
}

func (cnu CreateNewUser) SendEmail(email, name string, token []byte) error {
  host, addr, pass, port := GetValueEmail("./config.json", "noreply")
  auth := smtp.PlainAuth("", addr, pass, host)

  // Email Link should contain the email and token for later parsing
  template := GetValue("./jsonFiles/template.json", "VerifyUser")

  link, _ := url.Parse("https://localhost:8080/VerifyUser?")
  q := link.Query()
  q.Set("Token", hex.EncodeToString(token)) // Change First from HEX to String
  link.RawQuery = q.Encode()

  msg := []byte(cnu.ComposeMessage(template, name, link.String()))

  err := smtp.SendMail(host + ":" + port, auth, addr, []string{email}, msg)
  if err != nil {
    log.Fatal(err)
    return err
  }
  return nil
}

func (cnu CreateNewUser) ComposeMessage(template, name, link string) string {
  r := strings.NewReplacer("${LINK}", link, "${NAME}", name)
  return r.Replace(template)
}

type VerifyUser struct {
  db *sql.DB
}

func(vu *VerifyUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  queryString, err := url.ParseQuery(r.URL.RawQuery)
  if err != nil {
    log.Fatal("FAILED TO PARSE QUERY STRING:", err)
    w.WriteHeader(http.StatusBadRequest)
  }

  query := GetValue("./jsonFiles/query.json", "VerifyUser")
  _, err = ReadQuery(vu.db, query, queryString["token"])
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
  } else {
    w.WriteHeader(http.StatusOK)
  }
}

type LoginUser struct {
  db             *sql.DB
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

  query := GetValue("./jsonFiles/query.json", "LoginCredentials")
  _, err = ReadQuery(lu.db, query, login.Email + ":" + login.Password, lu.aesCredentials)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
  }
  expiration :=  time.Now().Add(30 * 24 * time.Hour)
  cookieToken := lu.CookieValue(login.Email, expiration)
  cookieValue := hex.EncodeToString(cookieToken)
  cookie := http.Cookie{Name: login.Email, Value: cookieValue, Expires: expiration}
  err = lu.SendCookieToDB(cookie)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
  }
  http.SetCookie(w, &cookie)
}

func (lu *LoginUser) CookieValue(email string, expiration time.Time) []byte {
  byteVersion := string(expiration.UnixNano())

  newToken := []byte(byteVersion + email)

  hash := sha512.New()
  hash.Write(newToken)

  return hash.Sum(nil)
}

func (lu *LoginUser) SendCookieToDB(cookie http.Cookie) error {
  query := GetValue("./jsonFiles/query.json", "RegisterCookie")
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

  query := GetValue("./jsonFiles/query.json", "ForgetEmail")
  result, err := ReadQuery(fp.db, query, email)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
  }
  result = result[0].([]interface{})
  err = fp.CreateToken(result[0].(string), result[1].(string))
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }
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
  query := GetValue("./jsonFiles/query.json", "CreateTokenForgetPass")
  err = DatabaseInsert(fp.db, query, email, timeString, token)
  if err != nil {
    return err
  }
  return nil
}

func (fp *ForgetPass) SendEmail(email, name string, token []byte) error {
  host, addr, pass, port := GetValueEmail("./config.json", "noreply")
  auth := smtp.PlainAuth("", addr, pass, host)

  // Email Link should contain the email and token for later parsing
  template := GetValue("./jsonFiles/template.json", "ForgetPass")

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

  query := GetValue("./jsonFiles/query.json", "VerifyEmail")
  _, err = ReadQuery(vt.db, query, queryString["email"])
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
  } else {
    w.WriteHeader(http.StatusOK)
  }

}

type PasswordRecovery struct {
  db             *sql.DB
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

  query := GetValue("./jsonFiles/query.json", "UpdatePassword")
  err = DatabaseInsert(pr.db, query, profile.Email + ":" + profile.Password, pr.aesCredentials)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }

  query = GetValue("./jsonFiles/query.json", "DeleteToken")
  err = DatabaseInsert(pr.db, query, profile.Email + ":" + profile.Password, pr.aesCredentials)
  if err != nil {
    log.Fatal(err)
    w.WriteHeader(http.StatusNotFound)
  }
  w.WriteHeader(http.StatusOK)
}
