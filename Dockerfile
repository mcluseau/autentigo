from mcluseau/golang-builder:1.12.7 as build-env

from alpine:3.9
entrypoint ["/bin/autentigo"]
copy --from=build-env /go/bin/ /bin/
