CREATE TABLE `user`
(
    `id`         int unsigned NOT NULL AUTO_INCREMENT,
    `nickname`   varchar(32) NOT NULL DEFAULT '' COMMENT '用户昵称',
    `api_key`    varchar(32) NOT NULL DEFAULT '' COMMENT 'APIKEY',
    `created_at` datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY          `activity2_id_score` (`activity2_id`,`score`)
) ENGINE=InnoDB DEFAULT AUTO_INCREMENT=106000  COMMENT='用户表';

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
    KEY          `user_id` (`user_id`),
) ENGINE=InnoDB  COMMENT='用户主机关系';

CREATE TABLE `user_files`
(
    `id`           int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`      int unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `proxy_id`     int unsigned NOT NULL DEFAULT 0 COMMENT '代理ID',
    `meta_id`      varchar(32)  NOT NULL DEFAULT '' COMMENT '请求路径hash',
    `origin_file`  varchar(255) NOT NULL DEFAULT '' COMMENT '原图在本地位置',
    `convert_file` varchar(255) NOT NULL DEFAULT '' COMMENT '转换后在本地位置',
    `origin_size`  int unsigned NOT NULL DEFAULT 0 COMMENT '请求地址',
    `convert_size` int unsigned NOT NULL DEFAULT 0 COMMENT '请求地址',
    `created_at`   datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY            `user_id_proxy_id_created_at` (`user_id`,`proxy_id`,`created_at`)
) ENGINE=InnoDB  COMMENT='用户主机关系';


CREATE TABLE `request_log`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`     int unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `proxy_id`    int unsigned NOT NULL DEFAULT '0' COMMENT '代理ID',
    `local_id`    varchar(32)  NOT NULL DEFAULT '' COMMENT '请求路径hash',
    `remote_path` varchar(255) NOT NULL DEFAULT '' COMMENT '请求地址',
    `referer`     varchar(255) NOT NULL DEFAULT '' COMMENT '请求头referer',
    `ip`          varchar(64)  NOT NULL DEFAULT '' COMMENT '请求ip',
    `is_cache`    tinyint(1) NOT NULL DEFAULT '0' COMMENT '1.缓存文件，0.溯源',
    `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY           `user_id_proxy_id_created_at` (`user_id`,`proxy_id`,`created_at`)
) ENGINE=InnoDB  COMMENT='用户主机关系';

CREATE TABLE `request_stat`
(
    `id`            int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`       int unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
    `proxy_id`      int unsigned NOT NULL DEFAULT '0' COMMENT '代理ID',
    `path_id`       varchar(32)  NOT NULL DEFAULT '' COMMENT '请求路径hash（空表示所有路径）',
    `path`          varchar(255) NOT NULL DEFAULT '' COMMENT '请求地址（空表示所有路径）',
    `request_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '请求次数',
    `response_byte` bigint unsigned NOT NULL DEFAULT '0' COMMENT '相应数据大小',
    `saved_byte`    bigint unsigned NOT NULL DEFAULT '0' COMMENT '减少消耗流量大小',
    `created_at`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY             `user_id_proxy_id_path_id` (`user_id`,`proxy_id`,`path_id`)
) ENGINE=InnoDB  COMMENT='请求统计';


