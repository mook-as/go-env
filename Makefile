all:	bin/go-env
	@echo "Launching at http://localhost:5050/"
	foreman start -p 5050

bin/go-env:
	GOBIN=${PWD}/bin CGO_ENABLED=0 go install -ldflags '-w -s' -installsuffix 'no-cgo'

docker: Dockerfile bin/go-env
	docker build -t mookas/go-env .
