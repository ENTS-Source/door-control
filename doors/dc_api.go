package doors

import (
	"log"
	"net/http"

	"github.com/ents-source/door-control/api/auth"
)

func InstallApi() {
	http.HandleFunc("/v1/doors/open", auth.Require(doOpenDoors))
}

func doOpenDoors(w http.ResponseWriter, r *http.Request) {
	for _, d := range All() {
		if err := d.Open(); err != nil {
			log.Println("Error opening door: ", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}
