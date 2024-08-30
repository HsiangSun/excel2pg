# 使用最小的 Alpine 镜像作为基础镜像
FROM alpine:latest

# 安装证书（如果你的应用需要与 HTTPS API 进行通信）
#RUN apk --no-cache add ca-certificates

RUN export GIN_MODE=release

# 设置工作目录
WORKDIR /app/

# 将本地编译好的二进制文件复制到容器中
COPY excel2pg .

# 将 .env 文件复制到容器中（可选）
COPY .env .env
COPY static static
COPY uploads uploads

# 暴露应用使用的端口（假设使用 7777）
EXPOSE 7777

# 启动应用
CMD ["./excel2pg"]
