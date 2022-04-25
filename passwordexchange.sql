CREATE DATABASE IF NOT EXISTS `passwordExchange`;
USE `passwordExchange`;
CREATE TABLE `messages` (
  `messageid` int(11) NOT NULL AUTO_INCREMENT,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `lastAccessed` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `viewed` int(11) DEFAULT '0',
  `firstname` varchar(100) DEFAULT NULL,
  `lastname` varchar(100) DEFAULT NULL,
  `other_firstname` varchar(100) DEFAULT NULL,
  `other_lastname` varchar(100) DEFAULT NULL,
  `message` text NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `other_email` varchar(255) DEFAULT NULL,
  `uniqueid` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`messageid`),
  UNIQUE KEY `uniqueid` (`uniqueid`)
) ENGINE=InnoDB;

CREATE USER IF NOT EXISTS 'passwordexchange'@'%' IDENTIFIED BY PASSWORD 'PASSWORD' ;
GRANT USAGE to 'passwordexchange'@'%' on passwordExchange.*;
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, INDEX, ALTER, CREATE TEMPORARY TABLES, LOCK TABLES ON `passwordexchange`.* TO 'passwordexchange'@'%' ;
CREATE USER IF NOT EXISTS 'deletemessages'@'%' IDENTIFIED BY PASSWORD 'PASSWORD'; 
GRANT USAGE, SELECT (created), DELETE ON `passwordExchange`.`messages` TO 'deletemessages'@'%' 