package proxy

import (
	"net/http"
)

func BasicAuthIfSet(req *http.Request, username string, password string) {
	if len(username) > 0 {
		req.SetBasicAuth(username, password)
	}
}
