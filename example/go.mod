module exactpro/th2/example

go 1.16

replace github.com/th2-net/th2-common-go => ../

require (
	github.com/google/uuid v1.1.2
	github.com/rabbitmq/amqp091-go v1.5.0 // indirect
	github.com/rs/zerolog v1.26.1
	github.com/th2-net/th2-common-go v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.51.0
	google.golang.org/protobuf v1.27.1
)
