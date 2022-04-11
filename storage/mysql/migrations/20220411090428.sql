CREATE TABLE `top_sites` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `title` varchar(150) DEFAULT NULL,
  `url` varchar(100) DEFAULT NULL,
  `description` text,
  `counter_id` int(11) DEFAULT NULL,
  `rubric_id` bigint(20) DEFAULT NULL,
  `visitors` int(11) DEFAULT 0,
  `hits` int(11) DEFAULT 0,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `show_digits` tinyint(1) DEFAULT 1,
  `active` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `visitors` (`visitors`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8