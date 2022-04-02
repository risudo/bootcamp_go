CREATE TABLE `account` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL UNIQUE,
  `password_hash` varchar(255) NOT NULL,
  `display_name` varchar(255),
  `avatar` text,
  `header` text,
  `note` text,
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
);

CREATE TABLE `status` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `account_id` bigint(20) NOT NULL,
  `content` text NOT NULL,
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_account_id` (`account_id`),
  CONSTRAINT `fk_status_account_id` FOREIGN KEY (`account_id`) REFERENCES  `account` (`id`)
);

CREATE TABLE `relation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `following_id` bigint(20) NOT NULL,
  `follower_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_following_id` (`following_id`),
  INDEX `idx_follower_id` (`follower_id`),
  CONSTRAINT `fk_follows_following_id` FOREIGN KEY (`following_id`) REFERENCES  `account` (`id`),
  CONSTRAINT `fk_follows_follower_id` FOREIGN KEY (`follower_id`) REFERENCES  `account` (`id`)
);

CREATE TABLE `attachment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `type` varchar(255) NOT NULL,
  `url` text NOT NULL,
  `descriptions` text,
  PRIMARY KEY (`id`)
);
