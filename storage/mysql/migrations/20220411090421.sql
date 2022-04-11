CREATE TABLE `top_data` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL,
  `sess_id` char(16) NOT NULL,
  `page` varchar(255) NOT NULL,
  `refferer` text NOT NULL,
  `ua` varchar(255) NOT NULL,
  `ip` char(15) NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `day` int(10) unsigned NOT NULL,
  `country` char(3) NOT NULL,
  `city` int(8) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `date` (`date`),
  KEY `sess_id` (`sess_id`),
  KEY `user_id_day` (`user_id`,`day`,`date`),
  KEY `day` (`day`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8