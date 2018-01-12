all:	bin/go-env
	@echo "Launching at http://localhost:5050/"
	foreman start -p 5050

# -trimpath doesn't seem to take multiple arguments.
TRIM_PATH_ARGS=-trimpath ${HOME}
bin/go-env: env.go
	GOBIN=${CURDIR}/bin CGO_ENABLED=0 go install -x -a -gcflags '${TRIM_PATH_ARGS}' -asmflags '${TRIM_PATH_ARGS}' -ldflags '-w -s' -installsuffix 'no-cgo'

docker: Dockerfile bin/go-env
	docker build -t mookas/go-env .
