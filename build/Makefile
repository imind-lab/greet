GOPATH := $(shell go env GOPATH)
VERSION := 0.0.1.1

gengo:
	protoc -I. --proto_path ../application/greeter/proto --proto_path ../pkg/proto\
 --go_out ../application/greeter/proto --go_opt paths=source_relative --go-grpc_out ../application/greeter/proto --go-grpc_opt paths=source_relative \
 --grpc-gateway_out ../application/greeter/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=false greeter.proto
	protoc-go-inject-tag -input=../application/greeter/proto/greeter.pb.go

depend: gengo
	go get ../...

build: depend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o greeter ../main.go

test:
	go test -v ../... -cover

docker: gengo
	docker build -f ./Dockerfile -t registry.cn-beijing.aliyuncs.com/imind/greeter:$(VERSION) ../
	docker push registry.cn-beijing.aliyuncs.com/imind/greeter:$(VERSION)

deploy: docker
	@helm upgrade --install reviews ../deploy/helm/greeter --set image.tag=$(VERSION)

clean:
	docker rmi registry.cn-beijing.aliyuncs.com/imind/greeter:$(VERSION)

k8s: docker
	kubectl set image deployment/greeter greeter=registry.cn-beijing.aliyuncs.com/imind/greeter:$(VERSION)

.PHONY: gengo depend build test docker deploy clean k8s
