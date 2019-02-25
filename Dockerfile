from golang:1.11.5-alpine3.9 as build-env
run apk update && apk add gcc musl-dev
env pkg github.com/mcluseau/autentigo
workdir /go/src/$pkg
add . .
run go test ./... \
 && go install . ./cmd/...

from alpine:3.9
entrypoint ["/bin/autentigo"]
copy --from=build-env /go/bin/ /bin/
