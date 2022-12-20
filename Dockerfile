FROM golang:1.19
RUN mkdir /app
ADD . /app
WORKDIR /app

# Installing protobuf compiler
RUN apt-get update
RUN apt-get install --no-install-recommends --assume-yes protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# Generating go code from proto
RUN proto/gen_proto.sh

# Building and running the code
RUN go build -o main .
CMD ["/app/main"]