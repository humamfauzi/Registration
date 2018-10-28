CREATE DATABASE IF NOT EXISTS registration;

USE registration;

CREATE TABLE IF NOT EXISTS profile_table (
    profile_id BIGINT AUTO_INCREMENT,
    profile_date DATE,
    profile_name VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    IS_TEST bool,
    PRIMARY KEY (profile_id),
    UNIQUE KEY (email)
	);

CREATE TABLE IF NOT EXISTS pass (
   email VARCHAR(255) NOT NULL,
   password BLOB NOT NULL,
   IS_TEST bool
   );
