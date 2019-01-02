package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"

	companionapi "github.com/mcluseau/autorizo/companion-api/api"
	"github.com/mcluseau/autorizo/companion-api/backend"
	"github.com/mcluseau/autorizo/companion-api/backend/etcd"
	"github.com/mcluseau/autorizo/companion-api/backend/users-file"
)

var (
	bind        = flag.String("bind", ":8181", "HTTP bind specification")
	disableCORS = flag.Bool("no-cors", false, "Disable CORS support")
)

func main() {
	flag.Parse()

	cAPI := &companionapi.CompanionAPI{
		Client: getBackEndClient(),
	}

	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.DefaultContainer.Router(restful.CurlyRouter{})

	restful.Add(cAPI.Register())

	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath:     "/apidocs.json",
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	if !*disableCORS {
		restful.Filter(restful.CrossOriginResourceSharing{
			CookiesAllowed: true,
			Container:      restful.DefaultContainer,
		}.Filter)
	}

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

	log.Fatal(http.Serve(l, restful.DefaultContainer))
}

func getBackEndClient() backend.Client {

	switch v := os.Getenv("AUTH_BACKEND"); v {
	case "stupid":
		log.Fatal("Stupid backend does not need the companion-api")
		return nil
	case "ldap-bind":
		log.Fatal("Please feel free to use a ldap client instead of the companion-api")
		return nil
	case "file":
		return usersfile.New(requireEnv("AUTH_FILE", "File containings users when using file auth"))
	case "etcd":
		return etcd.New(
			requireEnv("ETCD_PREFIX", "etcd prefix"),
			strings.Split(requireEnv("ETCD_ENDPOINTS", "etcd endpoints"), ","))
	default:
		log.Fatal("Unknown authenticator: ", v)
		return nil
	}
}

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}
