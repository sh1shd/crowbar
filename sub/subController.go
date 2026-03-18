package sub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SUBController handles HTTP requests for subscription links and JSON configurations.
type SUBController struct {
	subCustomHeaders string
	subCustomHtml   string
	subPath         string
	subEncrypt      bool

	subService     *SubService
}

// NewSUBController creates a new subscription controller with the given configuration.
func NewSUBController(
	g *gin.RouterGroup,
	subPath string,
	encrypt bool,
	rModel string,
	subCustomHeaders string,
	subCustomHtml string,
) *SUBController {
	sub := NewSubService(rModel)
	a := &SUBController{
		subCustomHeaders: subCustomHeaders,
		subCustomHtml:    subCustomHtml,
		subPath:          subPath,
		subEncrypt:       encrypt,

		subService:     sub,
	}
	a.initRouter(g)
	return a
}

// initRouter registers HTTP routes for subscription links and JSON endpoints
// on the provided router group.
func (a *SUBController) initRouter(g *gin.RouterGroup) {
	gLink := g.Group(a.subPath)
	gLink.GET(":subid", a.subs)
}

// subs handles HTTP requests for subscription links, returning either HTML page or base64-encoded subscription data.
func (a *SUBController) subs(c *gin.Context) {
	subId := c.Param("subid")
	scheme, host, hostWithPort, hostHeader := a.subService.ResolveRequest(c)
	subs, lastOnline, traffic, err := a.subService.GetSubs(subId, host)

	if err != nil || len(subs) == 0 {
		c.String(http.StatusBadRequest, "400 bad request")
	} else {
		result := ""
		for _, sub := range subs {
			result += sub + "\n"
		}

		// If the request was made from a web browser (by a accept header and user agent), render the info page here
		if strings.Contains(c.GetHeader("Accept"), "text/html") || strings.Contains(c.GetHeader("User-Agent"), "mozilla") {
			// Build page data in service
			subURL := a.subService.BuildURLs(scheme, hostWithPort, a.subPath, subId)
			// Get base_path from context (set by middleware)
			basePath, exists := c.Get("base_path")
			if !exists {
				basePath = "/"
			}
			// Add subId to base_path for asset URLs
			basePathStr := basePath.(string)
			if basePathStr == "/" {
				basePathStr = "/" + subId + "/"
			} else {
				// Remove trailing slash if exists, add subId, then add trailing slash
				basePathStr = strings.TrimRight(basePathStr, "/") + "/" + subId + "/"
			}
			page := a.subService.BuildPageData(subId, hostHeader, traffic, lastOnline, subs, subURL, basePathStr)

			// If custom HTML provided in settings, parse and execute it as a template
			if a.subCustomHtml != "" {
				tpl, err := template.New("sub_custom").Parse(a.subCustomHtml)
				if err == nil {
					_ = tpl.Execute(c.Writer, page)
					return
				}
			} else {
				// Fallback: minimal HTML output if custom template not available/failed
				fallbackTpl, err := template.New("sub_fallback").Parse(`
					<!DOCTYPE html>
					<html lang="en">
					<head>
						<meta charset="UTF-8">
						<meta name="viewport" content="width=device-width, initial-scale=1.0">
						<title>Subscription "{{.SId}}"</title>
					</head>
					<body>
						<h2>Client information</h2>
						<p>Subscription ID: {{.SId}}</p>
						<p>Download: {{.Download}}</p>
						<p>Upload: {{.Upload}}</p>
						<p>Used: {{.Used}}</p>
						<p>Total: {{.Total}}</p>
						{{if .SubUrl}}<p>URL: <a href="{{.SubUrl}}">{{.SubUrl}}</a></p>{{end}}
					</body>
					</html>
				`)
				if err == nil {
					c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
					_ = fallbackTpl.Execute(c.Writer, page)
					return
				}
			}
		}

		// Add headers
		subUserInfoHeader := fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", traffic.Up, traffic.Down, traffic.Total, traffic.ExpiryTime/1000)
		a.ApplyCommonHeaders(c, subUserInfoHeader, a.subCustomHeaders)  

		if a.subEncrypt {
			c.String(http.StatusOK, base64.StdEncoding.EncodeToString([]byte(result)))
		} else {
			c.String(http.StatusOK, result)
		}
	}
}

// ApplyCommonHeaders sets common HTTP headers for subscription responses including user info, and custom headers.
func (a *SUBController) ApplyCommonHeaders(
	c *gin.Context,
	userInfoHeader string,
	customHeadersJSON string,
) {
	// Parse and apply custom headers
	var customHeaders []map[string]string
	if customHeadersJSON != "" {
		if err := json.Unmarshal([]byte(customHeadersJSON), &customHeaders); err == nil {
			for _, hdr := range customHeaders {
				if name := hdr["name"]; name != "" {
					value := hdr["value"]

					c.Writer.Header().Set(name, value)
				}
			}
		}
	}

	// If the 'Profile-Update-Interval' header is missing, add it with a default value of 12 hours
	if c.Writer.Header().Get("Profile-Update-Interval") == "" {
		c.Writer.Header().Set("Profile-Update-Interval", "12")
	}

	// If the 'Profile-Title' header is missing, add it with a default value
	if c.Writer.Header().Get("Profile-Title") == "" {
		c.Writer.Header().Set("Profile-Title", "A Subscription Server")
	}

	c.Writer.Header().Set("Subscription-Userinfo", userInfoHeader)
}