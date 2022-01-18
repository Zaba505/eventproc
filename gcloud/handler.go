package gcloud

import (
	"net/http"
	"regexp"

	"github.com/Zaba505/eventproc/event"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

var router = mux.NewRouter()

func init() {
	viper.AddRemoteProvider("firestore", "google-cloud-project-id", "collection/document")
	viper.SetConfigType("json") // Config's format: "json", "toml", "yaml", "yml"
	err := viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile("^hello$")

	viper.Set("payloadRegex", map[*regexp.Regexp]string{
		re: "helloEndpoint",
	})

	router.
		Methods(http.MethodPost).
		Path("/event").
		Handler(event.NewHTTPHandler(nil))
}

func ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.ServeHTTP(w, req)
}
