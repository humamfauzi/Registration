# Deployment and Test
We wrap all Go files to an exe or any executable. We deploy it in a containerized
form. We also deploy database MySql 5.7 with container and an SMTP server for our mailing test.
So in total, there are four container which is our main app, database, and a SMTP server and IMAP server.

Our main app should be able to connect with database and SMTP server where we send mail.
We still don't know how to orchestrate a connection. We did have some experience setting
a containerized database but we did not have one for mail server. 

# Database Design Documentation

1. Profile Table
```
  CREATE TABLE IF NOT EXISTS profile (
    profile_id BIGINT AUTO_INCREMENT,
    date DATE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    PRIMARY KEY profile_id,
    UNIQUE KEY email
    )
```

2. Password Table
```
 CREATE TABLE IF NOT EXISTS pass (
   email VARCHAR(255) NOT NULL,
   password BLOB NOT NULL
   )
```

3. Forget Table
```
  CREATE TABLE IF NOT EXISTS token_list (
    date DATETIME,
    token BLOB NOT NULL,
    email VARCHAR(255) NOT NULL
    )
```
