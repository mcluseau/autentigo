from golang:1.12.1-alpine3.9 as build-env
run apk add --update git

env CGO_ENABLED=0
arg GOPROXY

workdir /src
add go.mod go.sum ./
run go mod download

add . ./
run go test ./...
run go install . ./cmd/...

from alpine:3.9
entrypoint ["/bin/autentigo"]
copy --from=build-env /go/bin/ /bin/
