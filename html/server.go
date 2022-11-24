// html build the Oware Web REST server
package html

// Web server

import (
	"fmt"
	"net/http"
	"sankofa/mech"
	"sankofa/ow"
)

// callback for web server;
// incorrect requests are redirected to the initial position
func PlayHandler(writer http.ResponseWriter, reader *http.Request) {
	rest := reader.URL.String()

	if rest == "/favicon.ico" {
		ow.Log("we don't serve hot icons")
		http.NotFound(writer, reader)
		return
	}

	if rest == "/" {
		initial := "/" + ow.Thousands(mech.INIRANK)
		ow.Log("redirecting empty to:", initial)
		http.Redirect(writer, reader, initial, 301)
		return
	}

	fmt.Fprint(writer, Display(rest))
	return
}
