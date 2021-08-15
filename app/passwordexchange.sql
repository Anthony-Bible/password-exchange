CREATE DATABASE IF NOT EXISTS `passwordExchange`;
USE `passwordExchange`;
CREATE TABLE IF NOT EXISTS `messages`
(
    messageid int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    lastAccessed TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    viewed INT DEFAULT 0,
    firstname varchar(100) NOT NULL,
    lastname varchar(100) NOT NULL,
    other_firstname varchar(100) NOT NULL,
    other_lastname varchar(100) NOT NULL,
    message text NOT NULL,
    email varchar(255) NOT NULL,
    other_email varchar(255) NOT NULL

);  
CREATE USER IF NOT EXISTS 'passwordexchange'@'%' IDENTIFIED BY PASSWORD 'PASSWORD' ;
GRANT USAGE to 'passwordexchange'@'%' on passwordExchange.*;