from golang:1.11.5-alpine3.8 as build-env
run apk update && apk add gcc musl-dev
env pkg github.com/mcluseau/autorizo
add . /go/src/$pkg
run cd /go/src/$pkg \
 && go test ./... \
 && go install . ./cmd/...

from alpine:3.8
entrypoint ["/bin/autorizo"]
copy --from=build-env /go/bin/ /bin
