# Deployment and Test
Untested Database

# Database Design Documentation

1. Profile Table
```
  CREATE TABLE IF NOT EXISTS profile (
    profile_id BIGINT AUTO_INCREMENT,
    date DATE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    IS_TEST bool,
    PRIMARY KEY profile_id,
    UNIQUE KEY email
    )
```

2. Password Table
```
 CREATE TABLE IF NOT EXISTS pass (
   email VARCHAR(255) NOT NULL,
   password BLOB NOT NULL
   IS_TEST bool,
   )
```

3. Forget Table
```
  CREATE TABLE IF NOT EXISTS token_list (
    date DATETIME,
    token BLOB NOT NULL,
    email VARCHAR(255) NOT NULL
    IS_TEST bool,
    )
```

4. Cookie Table
```
  CREATE TABLE IF NOT EXISTS cookie (
    expiration DATETIME,
    cookie VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL
    IS_TEST bool,
    )
```
