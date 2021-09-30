GOPATH := $(shell go env GOPATH)
VERSION := 0.0.1.1

gengo:
	protoc -I. --proto_path ../server/proto \
 --go_out ../server/proto --go_opt paths=source_relative --go-grpc_out ../server/proto --go-grpc_opt paths=source_relative \
 --grpc-gateway_out ../server/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=false greet/greet.proto
	protoc-go-inject-tag -input=../server/proto/greet/greet.pb.go

genphp:
	protoc -I. --proto_path ../server/proto \
 --php_out ../server/proto/greet --grpc_out ../server/proto/greet --plugin=protoc-gen-grpc=${GOPATH}/bin/grpc_php_plugin greet/greet.proto

depend:
	go get ../...

build: depend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o greet ../main.go

test:
	go test -v ../... -cover

docker: build health
	docker build -f ./Dockerfile -t registry.cn-beijing.aliyuncs.com/imind/greet:$(VERSION) ../
	#docker push registry.cn-beijing.aliyuncs.com/imind/greet:$(VERSION)
	rm -rf greet grpc-health-probe

health:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o grpc-health-probe ../pkg/grpc-health-probe/main.go

deploy: docker
	kubectl apply -f ns-rbac.yaml
	kubectl apply -f greet.yaml

helm:
	helm install greet ./helm --set image.tag=$(VERSION)

clean:
	docker rmi registry.cn-beijing.aliyuncs.com/imind/greet:$(VERSION)

k8s: docker
	kubectl set image deployment/greet greet=registry.cn-beijing.aliyuncs.com/imind/greet:$(VERSION)

.PHONY: gengo genphp depend build test docker health deploy helm clean k8s
