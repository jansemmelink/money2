CREATE DATABASE IF NOT EXISTS `don8`;
GRANT ALL PRIVILEGES ON `don8`.* to 'don8'@'%' IDENTIFIED BY 'don8';

DROP TABLE IF EXISTS `transactions`;
DROP TABLE IF EXISTS `statements`;
DROP TABLE IF EXISTS `bank_accounts`;
DROP TABLE IF EXISTS `accounts`;

CREATE TABLE `accounts` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `type` VARCHAR(20) NOT NULL,
  UNIQUE KEY `account_id` (`id`),
  UNIQUE KEY `account_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `bank_accounts` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `account_id` VARCHAR(40) NOT NULL,
  `bank_name` VARCHAR(100) NOT NULL,
  `account_number` VARCHAR(100) NOT NULL,
  `branch_name` VARCHAR(100) DEFAULT NULL,
  `branch_code` VARCHAR(100) DEFAULT NULL,
  UNIQUE KEY `bank_account_id` (`id`),
  UNIQUE KEY `bank_account_bank_number` (`bank_name`,`account_number`),
  FOREIGN KEY (`account_id`) REFERENCES `accounts`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `statements` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `bank_account_id` VARCHAR(40) NOT NULL,
  `opening_date` DATETIME NOT NULL,
  `opening_balance` VARCHAR(20) NOT NULL,
  `closing_date` DATETIME NOT NULL,
  `closing_balance` VARCHAR(20) NOT NULL,
  UNIQUE KEY `statement_id` (`id`),
  UNIQUE KEY `unique_statement` (`bank_account_id`,`opening_date`,`closing_date`),
  FOREIGN KEY (`bank_account_id`) REFERENCES `bank_accounts`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `transactions` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `date` DATETIME DEFAULT NULL,
  `amount` VARCHAR(100) NOT NULL,
  `dt_account_id` VARCHAR(40) DEFAULT NULL,
  `ct_account_id` VARCHAR(40) DEFAULT NULL,
  `statement_id` VARCHAR(40) DEFAULT NULL,
  `statement_type` VARCHAR(200) DEFAULT NULL,
  `statement_code` VARCHAR(200) DEFAULT NULL,
  `statement_details` VARCHAR(200) DEFAULT NULL,
  `notes` VARCHAR(200) DEFAULT NULL,
  UNIQUE KEY `transaction_id` (`id`),
  FOREIGN KEY (`statement_id`) REFERENCES `statements`(`id`),
  FOREIGN KEY (`dt_account_id`) REFERENCES `accounts`(`id`),
  FOREIGN KEY (`ct_account_id`) REFERENCES `accounts`(`id`),
  KEY `transaction_date` (`date`,`statement_id`),
  KEY `dt_account` (`dt_account_id`,`date`),
  KEY `ct_account` (`ct_account_id`,`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
