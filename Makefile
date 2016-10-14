all:	bin/go-env
	@echo "Launching at http://localhost:5050/"
	foreman start -p 5050

TRIM_PATH_ARGS=$(foreach path,$(subst :, ,${GOPATH}),-trimpath ${path}) -trimpath ${GOROOT} -trimpath ${PWD}
bin/go-env:
	GOBIN=${PWD}/bin CGO_ENABLED=0 go install -a -gcflags '${TRIM_PATH_ARGS}' -asmflags '${TRIM_PATH_ARGS}' -ldflags '-w -s' -installsuffix 'no-cgo'

docker: Dockerfile bin/go-env
	docker build -t mookas/go-env .
