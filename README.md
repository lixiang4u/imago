# 图片压缩


## 部分环境配置命令参考

```code

# ubuntu22

snap refresh
snap install certbot --classic
snap install go --classic
snap install node --classic


# 安装libvips依赖
# https://github.com/davidbyttow/govips/blob/master/build/Dockerfile-ubuntu-20.04
apt-get update -y
apt-get -y --no-install-recommends install software-properties-common gpg-agent
apt-get -y --no-install-recommends install build-essential devscripts lsb-release dput wget git nvi
add-apt-repository -y ppa:tonimelisma/ppa
add-apt-repository -y ppa:strukturag/libheif
add-apt-repository -y ppa:strukturag/libde265
apt-get -y install --no-install-recommends libvips-dev


# 安装nginx
# https://openresty.org/en/linux-packages.html#ubuntu
sudo apt-get -y install --no-install-recommends wget gnupg ca-certificates lsb-release
wget -O - https://openresty.org/package/pubkey.gpg | sudo gpg --dearmor -o /usr/share/keyrings/openresty.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/openresty.gpg] http://openresty.org/package/ubuntu $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/openresty.list > /dev/null
sudo apt-get update
sudo apt-get -y install openresty


# 申请证书，wildcard需要添加DNS TXT记录（注意提示的name值需要删除尾部域名部分）
# acme.sh(https://github.com/acmesh-official/acme.sh)就是个垃圾，很多二级域名什么都申请不成功
certbot certonly --nginx  -d   imago.artools.cc --nginx-server-root /usr/local/openresty/nginx/conf/
certbot certonly --manual -d *.imago.artools.cc


# 下载安装mariadb数据库
mkdir -p /opt/mariadb && cd /opt/mariadb 
wget https://dlm.mariadb.com/3669327/MariaDB/mariadb-11.2.2/repo/ubuntu/mariadb-11.2.2-ubuntu-jammy-amd64-debs.tar
tar -xf mariadb-11.2.2-ubuntu-jammy-amd64-debs.tar
./mariadb-11.2.2-ubuntu-jammy-amd64-debs/setup_repository
apt-get update && apt-get install mariadb-server


# 创建数据库以及用户
CREATE USER 'imago'@'localhost' IDENTIFIED BY 'password';
CREATE DATABASE imago CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON imago.* TO 'imago'@'localhost';
FLUSH PRIVILEGES;


# 下载nsq并启动
mkdir -p /opt/nsq/ && cd /opt/nsq/
wget https://s3.amazonaws.com/bitly-downloads/nsq/nsq-1.2.1.linux-amd64.go1.16.6.tar.gz
tar -zxf nsq-1.2.1.linux-amd64.go1.16.6.tar.gz
/opt/nsq/nsq-1.3.0.linux-amd64.go1.21.5/bin/nsqlookupd &
/opt/nsq/nsq-1.3.0.linux-amd64.go1.21.5/bin/nsqd --lookupd-tcp-address=127.0.0.1:4160 &
#/opt/nsq/nsq-1.3.0.linux-amd64.go1.21.5/bin/nsqadmin --lookupd-http-address=127.0.0.1:4161



```



## 部分nginx配置参考

> 主域名配置

```code
    server {
        listen       443 ssl;
        server_name  imago.artools.cc;

        ssl_certificate      /etc/letsencrypt/live/imago.artools.cc/fullchain.pem;
        ssl_certificate_key  /etc/letsencrypt/live/imago.artools.cc/privkey.pem;

        ssl_session_cache    shared:SSL:1m;
        ssl_session_timeout  5m;

        ssl_ciphers  HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers  on;

        root /opt/repo/imago-h5/dist;

        location / {
            index  index.html index.htm;
            try_files $uri $uri/ /index.html;
        }

        location /api {
            rewrite ^/api/(.*)$ /$1 break;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header Host $http_host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header Connection "";
            proxy_http_version 1.1;

            proxy_pass http://127.0.0.1:8060;
        }

    }
```

> 泛域名配置

```code

    server {
        listen       443 ssl;
        server_name  *.imago.artools.cc;

        ssl_certificate      /etc/letsencrypt/live/imago.artools.cc-0001/fullchain.pem;
        ssl_certificate_key  /etc/letsencrypt/live/imago.artools.cc-0001/privkey.pem;

        ssl_session_cache    shared:SSL:1m;
        ssl_session_timeout  5m;

        ssl_ciphers  HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers  on;

        root /opt/repo/imago-h5/dist;

        #location / {
        #    index  index.html index.htm;
        #    try_files $uri $uri/ /index.html;
        #}

        location / {
            rewrite ^/(.*)$ /$1 break;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header Host $http_host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header Connection "";
            proxy_http_version 1.1;

            proxy_pass http://127.0.0.1:8020;
        }

    }
```


