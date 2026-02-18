package sub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/mhsanaei/3x-ui/v2/config"

	"github.com/gin-gonic/gin"
)

// SUBController handles HTTP requests for subscription links and JSON configurations.
type SUBController struct {
	subTitle        string
	subCustomHeaders string
	subCustomHtml   string
	subPath         string
	subJsonPath     string
	jsonEnabled     bool
	subEncrypt      bool
	updateInterval  string

	subService     *SubService
	subJsonService *SubJsonService
}

// NewSUBController creates a new subscription controller with the given configuration.
func NewSUBController(
	g *gin.RouterGroup,
	subPath string,
	jsonPath string,
	jsonEnabled bool,
	encrypt bool,
	showInfo bool,
	rModel string,
	update string,
	jsonFragment string,
	jsonNoise string,
	jsonMux string,
	jsonRules string,
	subTitle string,
	subCustomHeaders string,
	subCustomHtml string,
) *SUBController {
	sub := NewSubService(showInfo, rModel)
	a := &SUBController{
		subTitle:         subTitle,
		subCustomHeaders: subCustomHeaders,
		subCustomHtml:    subCustomHtml,
		subPath:          subPath,
		subJsonPath:      jsonPath,
		jsonEnabled:      jsonEnabled,
		subEncrypt:       encrypt,
		updateInterval:   update,

		subService:     sub,
		subJsonService: NewSubJsonService(jsonFragment, jsonNoise, jsonMux, jsonRules, sub),
	}
	a.initRouter(g)
	return a
}

// initRouter registers HTTP routes for subscription links and JSON endpoints
// on the provided router group.
func (a *SUBController) initRouter(g *gin.RouterGroup) {
	gLink := g.Group(a.subPath)
	gLink.GET(":subid", a.subs)
	if a.jsonEnabled {
		gJson := g.Group(a.subJsonPath)
		gJson.GET(":subid", a.subJsons)
	}
}

// subs handles HTTP requests for subscription links, returning either HTML page or base64-encoded subscription data.
func (a *SUBController) subs(c *gin.Context) {
	subId := c.Param("subid")
	scheme, host, hostWithPort, hostHeader := a.subService.ResolveRequest(c)
	subs, lastOnline, traffic, err := a.subService.GetSubs(subId, host)
	if err != nil || len(subs) == 0 {
		c.String(400, "Error!")
	} else {
		result := ""
		for _, sub := range subs {
			result += sub + "\n"
		}

		// If the request expects HTML (e.g., browser) or explicitly asked (?html=1 or ?view=html), render the info page here
		accept := c.GetHeader("Accept")
		if strings.Contains(strings.ToLower(accept), "text/html") || c.Query("html") == "1" || strings.EqualFold(c.Query("view"), "html") {
			// Build page data in service
			subURL, subJsonURL := a.subService.BuildURLs(scheme, hostWithPort, a.subPath, a.subJsonPath, subId)
			if !a.jsonEnabled {
				subJsonURL = ""
			}
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
			page := a.subService.BuildPageData(subId, hostHeader, traffic, lastOnline, subs, subURL, subJsonURL, basePathStr)

			// If custom HTML provided in settings, parse and execute it as a template
			if a.subCustomHtml != "" {
				tpl, err := template.New("sub_custom").Parse(a.subCustomHtml)
				if err == nil {
					_ = tpl.Execute(c.Writer, gin.H{
						"title":        "subscription.title",
						"cur_ver":      config.GetVersion(),
						"host":         page.Host,
						"base_path":    page.BasePath,
						"sId":          page.SId,
						"download":     page.Download,
						"upload":       page.Upload,
						"total":        page.Total,
						"used":         page.Used,
						"remained":     page.Remained,
						"expire":       page.Expire,
						"lastOnline":   page.LastOnline,
						"downloadByte": page.DownloadByte,
						"uploadByte":   page.UploadByte,
						"totalByte":    page.TotalByte,
						"subUrl":       page.SubUrl,
						"subJsonUrl":   page.SubJsonUrl,
						"result":       page.Result,
					})
					return
				}
			} else {
				// Fallback: minimal HTML output if custom template not available/failed
				c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(c.Writer, "<html><head><title>Subscription %s</title></head><body>", page.SId)
				fmt.Fprintf(c.Writer, "<h1>Subscription %s</h1>", page.SId)
				fmt.Fprintf(c.Writer, "<p>Download: %s, Upload: %s, Used: %s, Total: %s</p>", page.Download, page.Upload, page.Used, page.Total)
				if page.SubUrl != "" {
					fmt.Fprintf(c.Writer, "<p>URL: <a href=\"%s\">%s</a></p>", page.SubUrl, page.SubUrl)
				}
				fmt.Fprint(c.Writer, "</body></html>")
				return	
			}
		}

		// Add headers
		header := fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", traffic.Up, traffic.Down, traffic.Total, traffic.ExpiryTime/1000)
		a.ApplyCommonHeaders(c, header, a.updateInterval, a.subTitle, a.subCustomHeaders)  

		if a.subEncrypt {
			c.String(200, base64.StdEncoding.EncodeToString([]byte(result)))
		} else {
			c.String(200, result)
		}
	}
}

// subJsons handles HTTP requests for JSON subscription configurations.
func (a *SUBController) subJsons(c *gin.Context) {
	subId := c.Param("subid")
	_, host, _, _ := a.subService.ResolveRequest(c)
	jsonSub, header, err := a.subJsonService.GetJson(subId, host)
	if err != nil || len(jsonSub) == 0 {
		c.String(400, "Error!")
	} else {
		// Add headers
		a.ApplyCommonHeaders(c, header, a.updateInterval, a.subTitle, a.subCustomHeaders)

		c.String(200, jsonSub)
	}
}

// ApplyCommonHeaders sets common HTTP headers for subscription responses including user info, update interval, and custom headers.
func (a *SUBController) ApplyCommonHeaders(
	c *gin.Context,
	header,
	updateInterval,
	profileTitle string,
	customHeadersJSON string,
) {
	c.Writer.Header().Set("Subscription-Userinfo", header)
	c.Writer.Header().Set("Profile-Update-Interval", updateInterval)

	// Set Profile-Title header
	if profileTitle != "" {
		c.Writer.Header().Set("Profile-Title", "base64:"+base64.StdEncoding.EncodeToString([]byte(profileTitle)))
	}

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
}