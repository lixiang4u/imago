CREATE TABLE `user`
(
    `id`         int unsigned NOT NULL AUTO_INCREMENT,
    `nickname`   varchar(32) NOT NULL DEFAULT '' COMMENT '用户昵称',
    `email`      varchar(32) NOT NULL DEFAULT '' COMMENT '邮件地址',
    `password`   varchar(32) NOT NULL DEFAULT '' COMMENT '密码',
    `api_key`    varchar(32) NOT NULL DEFAULT '' COMMENT 'APIKEY',
    `created_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `email` (`email`),
    UNIQUE KEY `api_key` (`api_key`)
) ENGINE=InnoDB AUTO_INCREMENT=106000  COMMENT='用户表';

CREATE TABLE `user_proxy`
(
    `id`         int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`    int unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `title`      varchar(32)  NOT NULL DEFAULT '' COMMENT '备注名，标题',
    `origin`     varchar(64)  NOT NULL DEFAULT '' COMMENT '源站地址：https://img.google.com',
    `host`       varchar(64)  NOT NULL DEFAULT '' COMMENT '代理主机：xooloop.imago-service.xyz',
    `quality`    tinyint(1) NOT NULL DEFAULT '80' COMMENT '图片质量',
    `user_agent` varchar(64)  NOT NULL DEFAULT '' COMMENT '溯源用的User-Agent',
    `cors`       varchar(255) NOT NULL DEFAULT '*' COMMENT 'CORS头（Access-Control-Allow-Origin）',
    `referer`    varchar(255) NOT NULL DEFAULT '*' COMMENT '',
    `status`     tinyint(1) NOT NULL DEFAULT '1' COMMENT '1.正常，0.未开启',
    `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `host` (`host`),
    KEY          `user_id` (`user_id`)
) ENGINE=InnoDB  COMMENT='用户主机关系';

CREATE TABLE `user_files`
(
    `id`           int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`      int unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `proxy_id`     int unsigned NOT NULL DEFAULT 0 COMMENT '代理ID',
    `meta_id`      varchar(32)  NOT NULL DEFAULT '' COMMENT '请求路径hash',
    `origin_url`   varchar(255) NOT NULL DEFAULT '' COMMENT '原图url路径',
    `origin_file`  varchar(255) NOT NULL DEFAULT '' COMMENT '原图在本地位置',
    `convert_file` varchar(255) NOT NULL DEFAULT '' COMMENT '转换后在本地位置',
    `origin_size`  int unsigned NOT NULL DEFAULT 0 COMMENT '源文件，字节',
    `convert_size` int unsigned NOT NULL DEFAULT 0 COMMENT '转换后，字节',
    `created_at`   datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY            `user_id_proxy_id_created_at` (`user_id`,`proxy_id`,`created_at`),
    KEY            `meta_id` (`meta_id`)
) ENGINE=InnoDB  COMMENT='用户主机关系';


CREATE TABLE `request_log`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`     int unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `proxy_id`    int unsigned NOT NULL DEFAULT 0 COMMENT '代理ID',
    `meta_id`     varchar(32)  NOT NULL DEFAULT '' COMMENT '请求路径hash',
    `request_url` varchar(255) NOT NULL DEFAULT '' COMMENT '请求url路径',
    `origin_url`  varchar(255) NOT NULL DEFAULT '' COMMENT '原图url路径',
    `referer`     varchar(255) NOT NULL DEFAULT '' COMMENT '请求头referer',
    `ua`          varchar(255) NOT NULL DEFAULT '' COMMENT '请求头UA',
    `ip`          varchar(64)  NOT NULL DEFAULT '' COMMENT '请求ip',
    `is_cache`    tinyint(1) NOT NULL DEFAULT 0 COMMENT '1.缓存文件，0.溯源',
    `is_exist`    tinyint(1) NOT NULL DEFAULT 0 COMMENT '1.源文件存在，0.源文件不存在',
    `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY           `user_id_proxy_id_created_at` (`user_id`,`proxy_id`,`created_at`)
) ENGINE=InnoDB  COMMENT='用户主机关系';

CREATE TABLE `request_stat`
(
    `id`            int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`       int unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `proxy_id`      int unsigned NOT NULL DEFAULT 0 COMMENT '代理ID',
    `meta_id`       varchar(32)  NOT NULL DEFAULT '' COMMENT '请求路径hash',
    `origin_url`    varchar(255) NOT NULL DEFAULT '' COMMENT '原图url路径（空表示所有路径）',
    `request_count` bigint unsigned NOT NULL DEFAULT 0 COMMENT '请求次数',
    `response_byte` bigint unsigned NOT NULL DEFAULT 0 COMMENT '相应数据大小',
    `saved_byte`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '减少消耗流量大小',
    `created_at`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY             `user_id_proxy_id_meta_id` (`user_id`,`proxy_id`,`meta_id`)
) ENGINE=InnoDB  COMMENT='请求统计';


