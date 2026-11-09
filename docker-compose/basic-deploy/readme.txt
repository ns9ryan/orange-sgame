之前已部署：
==================
mysql
```
docker run \
-p 3306:3306 \
--restart=always \
--name mysql-server \
--privileged=true \
-v /home/repo/docker-data/mysql/log:/var/log/mysql \
-v /home/repo/docker-data/mysql/data:/var/lib/mysql \
-v /home/repo/docker-data/mysql/conf/my.cnf:/etc/mysql/my.cnf \
-e MYSQL_ROOT_PASSWORD=root \
-d mysql
```

redis
```
docker pull redis
mkdir /docker-data/redis
wget http://download.redis.io/redis-stable/redis.conf
vi /home/repo/docker-data/redis/redis.conf

bind 127.0.0.1 # 这行要注释掉，解除本地连接限制
protected-mode no # 默认yes，如果设置为yes，则只允许在本机的回环连接，其他机器无法连接。
daemonize no # 默认no 为不守护进程模式，docker部署不需要改为yes，docker run -d本身就是后台启动，不然会冲突
requirepass 123456 # 设置密码
appendonly yes # 持久化


docker run --name redis \
-p 6379:6379 \
-v /home/repo/docker-data/redis/redis.conf:/etc/redis/redis.conf \
-v /home/repo/docker-data/redis:/data \
-d redis redis-server /etc/redis/redis.conf
```

mongodb
```
docker pull mongodb/mongodb-community-server:latest
docker run --name mongodb -p 27017:27017 -d mongodb/mongodb-community-server:latest
```

最近部署：
====================
etcd后台管理地址：
http://127.0.0.1:8088

nats管理地址：
nats1-1：http://127.0.0.1:8222
nats2-1：http://127.0.0.1:8223
nats3-1：http://127.0.0.1:8224

游戏服务，主要包括三大类:
nodeid="game"       
游戏主体逻辑，包含大厅服务/longhu1控端处理服务/longhu1游戏玩家服务服务等，主端口127.0.0.1:8801

nodeid="gateCtrl"
控端网关，主要包括控端网关服务/http控端网关服务等，主端口127.0.0.1:8802，控端用户端口为9001

nodeid="gate"
玩家网关，主要包括玩家网关服务/http玩家网关服务等，主端口127.0.0.1:8803，玩家用户端口为9000