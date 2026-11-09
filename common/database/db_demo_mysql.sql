
CREATE DATABASE /*!32312 IF NOT EXISTS*/`db_video_game` /*!40100 DEFAULT CHARACTER SET utf8 */;

USE `db_video_game`;

/*Table structure for table `vg_users` */

DROP TABLE IF EXISTS `vg_users`;

CREATE TABLE `vg_users` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_name` VARCHAR(64) DEFAULT '' COMMENT '账号',
  `user_id` VARCHAR(64) DEFAULT '' COMMENT 'user_id',
  `email` VARCHAR(64) DEFAULT '' COMMENT '邮箱',
  `pass` VARCHAR(128) DEFAULT '' COMMENT '密码',
  `real_name` VARCHAR(64) DEFAULT '' COMMENT '姓名',
  `phone` CHAR(64) DEFAULT '' COMMENT '手机',
  `avatar` VARCHAR(255) DEFAULT '' COMMENT '头像',
  `status` TINYINT(4) DEFAULT 1 COMMENT '状态',
  `remark` VARCHAR(255) DEFAULT '' COMMENT '备注',
  `last_login_time` VARCHAR(64) DEFAULT '',
  `last_login_ip` CHAR(30) DEFAULT '' COMMENT '最近一次登录ip',
  `login_times` INT(11) DEFAULT 0 COMMENT '累计登录次数',
  `wallet` decimal(11, 2) DEFAULT 0.00 COMMENT '余额',
  `created_at` VARCHAR(64) DEFAULT '',
  `updated_at` VARCHAR(64) DEFAULT '',
  `deleted_at` VARCHAR(64) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

INSERT INTO `vg_users` (
    `user_name`, `user_id`, `email`, `pass`, `real_name`, `phone`, `status`,
    `remark`, `last_login_time`, `last_login_ip`, `login_times`, `created_at`, `updated_at`,`avatar`,`wallet`)
VALUES ('test', 'test_user_id','test@qq.com','87d9bb400c0634691f0e3baaf1e2fd0d', 'test', '12345678912', 1,'0', '2024-11-22 17:40:22', '127.0.0.1', 1, '2024-11-22 17:40:22', '2024-11-22 17:40:22','xxxxx',9999999);
