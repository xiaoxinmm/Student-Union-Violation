FROM alpine:3.19

WORKDIR /app

# Copy pre-built binary
COPY server .
COPY web/ ./web/

# Create uploads directory
RUN mkdir -p /app/uploads && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

EXPOSE 8080

CMD ["./server"]
