package Registration

import (
  "crypto/rsa"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"

  "encoding/json"
  "io/ioutil"
  "os"

  jwt "github.com/dgrijalva/jwt-go"
)

// JSON Web Token Support
// HMAC  -- HS256 HS384 HS512 --> []byte for signing and validation
// RSA   -- RS256 RS384 RS512 --> *rsa.PrivateKey for signing and *rsa.PublicKey for Validation
// ECDSA -- ES256 ES384 ES512 --> *ecdsa.PrivateKey for signing and *ecdsa.PublicKey for Validation

// If a utils function grows large, there will be division within utils function

const (
  privateKeyPath = "pairkeys/app.rsa"
  publicKeyPath  = "pairkeys/app.rsa.pub"
)

var (
  verifyKey *rsa.PublicKey
  signKey   *rsa.PrivateKey
)

func init() {
  // initialize RSA key pair
  signBytes, err := ioutil.ReadFile(privateKeyPath)
  if err != nil {
    log.Fatal(err)
  }

  signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
  if err != nil {
    log.Fatal(err)
  }

  verifyBytes, err := ioutil.ReadFile(publicKeyPath)
  if err != nil {
    log.Fatal(err)
  }

  verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
  if err != nil {
    log.Fatal(err)
  }
}

// -------------------------TOKEN SECTION-------------------------------------

func TokenSigning(user, accessLevel string) (string, error){
  t := jwt.New(jwt.GetSigningMethod("RS256"))

  claimsDict := jwt.MapClaims{}
  claimsDict["AccesToken"] = accessLevel
  claimsDict["CustomUserInfo"] = struct {
    Name string
    Kind string
  }{user, "human"}
  claimsDict["exp"] = time.Now().Add(time.Minute * 1).Unix()
  t.Claims = claimsDict

  tokenString, err := t.SignedString(signKey)
  if err != nil {
    return tokenString, err
  }

  return tokenString, nil
}

func TokenParsing(tokenCookie string) error {
  token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
    return verifyKey, nil
  })
  if err != nil{
    return err
  }
  return token
}

type Dictionary map[string]string


// -------------------------------FETCHING VALUE----------------------------------

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


// -----------------------------DATABASE SECTION------------------------------

func DatabaseInsert(db *sql.DB, query string, input ...interface{}) error {
  input = inputFilter(input...)

  stmt, err := db.Prepare(query)
  if err != nil {
    return err
  }

  rows, err := stmt.Exec(input...)
  if err != nil {
    return err
  }

  _, err = rows.LastInsertId()
  if err != nil {
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

func inputFilter(input ...interface{}) []interface{} {
  for i := 0; i < len(input); i++ {
    switch input[i].(type) {
    case int:
      if input[i] == 0 {
        var ph sql.NullInt64
        input[i] = ph
      }
    case float64:
      if input[i] == 0 {
        var ph sql.NullFloat64
        input[i] = ph
      }
    case string:
      if len(input[i].(string)) == 0{
        var ph sql.NullString
        input[i] = ph
      }
    }
  }
  return input
}
