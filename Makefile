.PHONY: build clean deploy

build: 
	cd postIndex && env GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/postIndex ./postIndex.go && cd ..
	cd postSearch && env GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/postSearch ./postSearch.go && cd ..

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
