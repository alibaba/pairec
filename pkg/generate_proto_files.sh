#!/bin/bash

set -ex
# install gRPC and protoc plugin for Go, see http://www.grpc.io/docs/quickstart/go.html#generate-grpc-code
#mkdir tensorflow tensorflow_serving
protoc -I generate_golang_files/ generate_golang_files/tensorflow_serving/apis/*.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative   --go_out=. --go_opt=paths=source_relative
protoc -I generate_golang_files/ generate_golang_files/tensorflow/core/framework/* --go-grpc_out=. --go-grpc_opt=paths=source_relative   --go_out=. --go_opt=paths=source_relative
protoc -I generate_golang_files/ generate_golang_files/tensorflow/core/example/* --go-grpc_out=. --go-grpc_opt=paths=source_relative   --go_out=. --go_opt=paths=source_relative
