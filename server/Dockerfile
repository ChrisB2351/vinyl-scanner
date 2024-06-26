FROM golang:1.22-alpine3.19 as build

WORKDIR /app/

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY . /app/
RUN go build -o vinyl-scanner .

FROM alpine:3.19

COPY --from=build /app/vinyl-scanner /bin/vinyl-scanner

ENV UID 501

RUN apk update && \
  apk add --no-cache ca-certificates tzdata && \
  addgroup -g $UID user && \
  adduser -D -h /app -u 1000 -G users user && \
  mkdir -p /data && chown -R user /data

USER user

WORKDIR /data
VOLUME /data

ENV VINYL_DATA_DIR "/data"

EXPOSE 8080

CMD ["vinyl-scanner"]
