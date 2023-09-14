module github.com/nats-io/nats-server/v2

go 1.20

require (
	github.com/klauspost/compress v1.16.7
	github.com/minio/highwayhash v1.0.2
	github.com/nats-io/jwt/v2 v2.5.0
	github.com/nats-io/nats.go v1.28.0
	github.com/nats-io/nkeys v0.4.4
	github.com/nats-io/nuid v1.0.1
	go.uber.org/automaxprocs v1.5.3
	golang.org/x/crypto v0.12.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
	golang.org/x/sys v0.12.0
	golang.org/x/time v0.3.0
)

require (
	github.com/golang/protobuf v1.4.2 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
)

replace github.com/nats-io/nats.go v1.28.0 => github.com/M4SS-Code/nats.go v1.28.1-stream-source-heartbeat
