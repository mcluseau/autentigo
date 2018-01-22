package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/mcluseau/autorizo/api"
	"github.com/mcluseau/autorizo/auth/stupid-auth"
)

var (
	bind = flag.String("bind", ":8080", "HTTP bind specification")
)

func main() {
	key, pubKey, sm := initJWT()

	hAPI := &api.API{
		Authenticator: stupidauth.New(),
		PrivateKey:    key,
		PublicKey:     pubKey,
		SigningMethod: sm,
	}

	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.DefaultContainer.Router(restful.CurlyRouter{})

	restful.Add(hAPI.Register())

	config := swagger.Config{
		WebServices: restful.DefaultContainer.RegisteredWebServices(),
		ApiPath:     "/apidocs.json",
	}
	swagger.InstallSwaggerService(config)

	l, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("listening on ", *bind)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Kill, os.Interrupt, syscall.SIGTERM)
		<-sig

		log.Print("closing listener")
		l.Close()
	}()

	if err := http.Serve(l, restful.DefaultContainer); err != nil {
		log.Fatal(err)
	}
}

func initJWT() (key interface{}, cert interface{}, method jwt.SigningMethod) {
	crtData := requireEnv("TLS_CRT", "certificate used to sign/verify tokens")
	keyData := requireEnv("TLS_KEY", "key used to sign tokens")
	sm := requireEnv("SIGNING_METHOD", "signature method to use (must match the key)")

	method = jwt.GetSigningMethod(sm)

	switch sm[:2] {
	case "RS":
		if x, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyData)); err == nil {
			key = x
		} else {
			log.Fatal("Failed to load private key: ", err)
		}
		if x, err := jwt.ParseRSAPublicKeyFromPEM([]byte(crtData)); err == nil {
			cert = x
		} else {
			log.Fatal("Failed to load public key: ", err)
		}

	case "ES":
		if x, err := jwt.ParseECPrivateKeyFromPEM([]byte(keyData)); err == nil {
			key = x
		} else {
			log.Fatal("Failed to load private key: ", err)
		}
		if x, err := jwt.ParseECPublicKeyFromPEM([]byte(crtData)); err == nil {
			cert = x
		} else {
			log.Fatal("Failed to load public key: ", err)
		}

	default:
		log.Fatal("Invalid SIGNING_METHOD: ", sm)
	}
	return
}

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}
