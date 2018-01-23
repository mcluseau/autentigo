from golang:1.9.2-alpine3.7 as build-env
env pkg github.com/mcluseau/autorizo
add . /go/src/$pkg
run cd /go/src/$pkg \
 && go vet ./... \
 && go test ./... \
 && go install .

from alpine:3.7
entrypoint ["/bin/autorizo"]
copy --from=build-env /go/bin/autorizo /bin
