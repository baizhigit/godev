package api

import _ "embed"

//go:embed user.swagger.json
var SwaggerJSON []byte // экспортируем — используем в httpserver
