from golang:1.11.1-alpine3.8 as build-env
env pkg github.com/mcluseau/autorizo
add . /go/src/$pkg
run cd /go/src/$pkg \
 && go vet ./... \
 && go test ./... \
 && go install .

from alpine:3.8
entrypoint ["/bin/autorizo"]
copy --from=build-env /go/bin/autorizo /bin
