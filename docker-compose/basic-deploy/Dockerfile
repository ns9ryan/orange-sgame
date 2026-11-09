# 使用的基础镜像
FROM ubuntu:22.04

# 设置时区
ENV TZ=Asia/Shanghai

# 更新包列表并安装必要的依赖项，包括 tzdata 来正确配置时区
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 默认命令，假设通过启动参数来启动不同的节点
CMD ["./orange-sgame"]