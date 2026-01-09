package handler

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SwaggerHandler struct{}

func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// ServeSwaggerJSON serves the generated swagger.json file
func (h *SwaggerHandler) ServeSwaggerJSON(c echo.Context) error {
	return c.File("./docs/swagger.json")
}

// ServeSwaggerUI serves the Swagger UI documentation interface via CDN
func (h *SwaggerHandler) ServeSwaggerUI(c echo.Context) error {
	html := `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Base Project API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
      body { margin: 0; padding: 0; }
      .swagger-ui .topbar { display: none; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        SwaggerUIBundle({
          url: "/swagger/doc.json",
          dom_id: "#swagger-ui",
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIBundle.SwaggerUIStandalonePreset
          ],
          layout: "BaseLayout",
          deepLinking: true,
          showExtensions: true,
          showCommonExtensions: true
        });
      };
    </script>
  </body>
</html>`

	tmpl := template.Must(template.New("swagger").Parse(html))
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return tmpl.Execute(c.Response().Writer, nil)
}
