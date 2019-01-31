from golang:1.11.5-alpine3.8 as build-env
run apk update && apk add gcc musl-dev
env pkg github.com/mcluseau/autorizo/companion-api
add . /go/src/$pkg

run cd /go/src/$pkg \
 && go vet ./... \
 && go test ./... \
 && go install .

from alpine:3.8
entrypoint ["/bin/companion-api"]
copy --from=build-env /go/bin/companion-api /bin
