CREATE TABLE IF NOT EXISTS `messages` (
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
);
