CREATE TABLE `top_dynamics` (
  `site_id` int(11) DEFAULT 0,
  `hosts` int(11) NOT NULL DEFAULT 0,
  `visitors` int(11) NOT NULL DEFAULT 0,
  `hits` int(11) NOT NULL DEFAULT 0,
  `date` date DEFAULT NULL,
  `yandex` int(11) NOT NULL DEFAULT 0,
  `google` int(11) NOT NULL DEFAULT 0,
  `mail` int(11) NOT NULL DEFAULT 0,
  `bing` int(11) NOT NULL DEFAULT 0,
  `rambler` int(11) NOT NULL DEFAULT 0,
  `date_ts` int(20) unsigned NOT NULL,
  `year` mediumint(4) unsigned NOT NULL,
  `month` mediumint(5) unsigned NOT NULL,
  UNIQUE KEY `site_id` (`site_id`,`date`),
  KEY `site_id_year` (`site_id`,`year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8