FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/fifa .

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=build /out/fifa ./fifa
COPY login.html index.html ./
RUN mkdir -p /data && chown appuser /data /app
USER appuser
ENV ADDR=:8080
ENV DB_PATH=/data/users.db
EXPOSE 8080
ENTRYPOINT ["./fifa"]
