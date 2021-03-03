CREATE TABLE `files` (
	`id` CHAR(20) NOT NULL COLLATE 'ascii_bin',
	`created_at` DATETIME NOT NULL DEFAULT current_timestamp(),
	`uploaded` TINYINT(1) NOT NULL DEFAULT '0',
	`filename` TEXT NULL DEFAULT NULL COLLATE 'utf8mb4_unicode_ci',
	`expires` DATETIME NULL DEFAULT NULL,
	`lock` INT(11) NOT NULL DEFAULT '0',
	PRIMARY KEY (`id`) USING BTREE
)
COLLATE='utf8mb4_unicode_ci'
ENGINE=InnoDB
;
