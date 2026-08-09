FROM kouleen/golang:1.25 AS builder
LABEL authors="Kouleen.china@gmail.com"
# 设置工作目录
WORKDIR /app
# 复制 go.mod 和 go.sum 文件，提前下载依赖
COPY go.mod go.sum ./
# 设置代理，加速下载
RUN #go env -w GOPROXY=https://goproxy.cn,direct && go mod download
RUN go mod download
# 复制整个项目到工作目录
COPY . .
# 构建应用，禁用 CGO
RUN CGO_ENABLED=0 go build -o lets-encrypt ./main.go

# 第二阶段：创建轻量级运行时镜像
FROM kouleen/alpine:latest
LABEL authors="Kouleen.china@gmail.com"
RUN apk add --no-cache tzdata && rm -rf /var/cache/apk/*

# 设置工作目录
WORKDIR /app
# 申明容器编码
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US:en \
    LC_ALL=en_US.UTF-8 \
    TZ=Asia/Shanghai
# 从第一阶段的镜像中复制构建好的应用
COPY --from=builder /app/lets-encrypt .

# 暴露应用监听的端口，根据实际应用修改
EXPOSE 8099

# 运行应用
CMD ["./lets-encrypt"]