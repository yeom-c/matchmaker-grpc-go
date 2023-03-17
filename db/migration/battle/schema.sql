SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for battle_result
-- ----------------------------
DROP TABLE IF EXISTS `battle_result`;
CREATE TABLE `battle_result` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_user_id` int(10) unsigned NOT NULL,
  `match_account_user_id` int(10) unsigned NOT NULL DEFAULT 0,
  `channel_id` varchar(255) NOT NULL DEFAULT '',
  `deck_id` int(10) unsigned NOT NULL DEFAULT 0,
  `result` int(5) unsigned NOT NULL DEFAULT 0,
  `battle_start_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `confirmed_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq_account_user_id_channel_id` (`account_user_id`,`channel_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for schema_migrations
-- ----------------------------
DROP TABLE IF EXISTS `schema_migrations`;
CREATE TABLE `schema_migrations` (
  `version` bigint(20) NOT NULL,
  `dirty` tinyint(1) NOT NULL,
  PRIMARY KEY (`version`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_user_id` int(10) NOT NULL,
  `match_point` int(10) unsigned NOT NULL DEFAULT 0,
  `match_win` int(10) unsigned NOT NULL DEFAULT 0,
  `match_lose` int(10) unsigned NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_account_user_id` (`account_user_id`) USING BTREE,
  KEY `idx_match_point` (`match_point`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4;

SET FOREIGN_KEY_CHECKS = 1;
