-- KWeaver Compose: extra databases beyond MYSQL_DATABASE bootstrap (official image creates MYSQL_USER on adp DB).
CREATE DATABASE IF NOT EXISTS sandbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON sandbox.* TO 'adp'@'%';
