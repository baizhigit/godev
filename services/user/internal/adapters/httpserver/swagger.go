package httpserver

import (
	"net/http"

	"github.com/baizhigit/godev/services/user/api"
)

func SwaggerHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(api.SwaggerJSON)
	})

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(scalarHTML)
	})

	return mux
}

var scalarHTML = []byte(`<!doctype html>
<html>
<head>
  <title>User Service — API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
</head>
<body>
  <script
    id="api-reference"
    data-url="/swagger.json"
    data-configuration='{"theme":"purple","layout":"modern","defaultHttpClient":{"targetKey":"shell","clientKey":"curl"}}'
  ></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`)
