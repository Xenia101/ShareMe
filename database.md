# Database

## files

|Column Name|Type|Description|
|-|:-:|-|
|`id`|`CHAR(32)`|id|
|`file_name`|`TEXT`|Original FileName|
|`expired`|`datetime`||
|`password`|`char(64)`|Password (sha-256)|

```sql
CREATE TABLE `files` (
    `id` CHAR(10) NOT NULL COLLATE 'ascii_bin',
    `filename` TEXT NOT NULL COLLATE 'utf8mb4_unicode_ci',
    `expires` DATETIME NOT NULL,
    `password` CHAR(64) NULL DEFAULT NULL COMMENT 'sha256' COLLATE 'ascii_bin',
    PRIMARY KEY (`id`) USING BTREE
)
COLLATE='utf8mb4_unicode_ci'
ENGINE=InnoDB
;
```
