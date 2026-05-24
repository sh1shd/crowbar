package minifier

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v2/config"
	"golang.org/x/net/html"
)

type captureWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *captureWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *captureWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func HTMLMinifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.IsDebug() {
			c.Next()
			return
		}

		writer := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		contentType := strings.ToLower(c.Writer.Header().Get("Content-Type"))
		origWriter := writer.ResponseWriter
		if !strings.Contains(contentType, "text/html") {
			if writer.status != 0 {
				origWriter.WriteHeader(writer.status)
			}
			origWriter.Write(writer.body.Bytes())
			return
		}

		minified, err := minifyInlineHTML(writer.body.Bytes())
		if err != nil {
			minified = writer.body.Bytes()
		}

		origWriter.Header().Del("Content-Length")
		if writer.status != 0 {
			origWriter.WriteHeader(writer.status)
		}
		origWriter.Write(minified)
	}
}

func minifyInlineHTML(content []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	combineInlineTags(doc, "script", isInlineJavaScript)
	combineAllInlineStyles(doc)
	walkHTMLNode(doc)

	var out bytes.Buffer
	if err := html.Render(&out, doc); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func combineInlineTags(root *html.Node, tag string, matcher func(*html.Node) bool) {
	var process func(parent *html.Node)
	process = func(parent *html.Node) {
		for node := parent.FirstChild; node != nil; {
			if node.Type == html.ElementNode && node.Data == tag && matcher(node) {
				group := []*html.Node{node}
				next := node.NextSibling
				for next != nil {
					if next.Type == html.TextNode && strings.TrimSpace(next.Data) == "" {
						group = append(group, next)
						next = next.NextSibling
						continue
					}
					if next.Type == html.ElementNode && next.Data == tag && matcher(next) {
						group = append(group, next)
						next = next.NextSibling
						continue
					}
					break
				}

				if len(group) > 1 {
					combineGroup(parent, tag, group)
					node = group[len(group)-1].NextSibling
					continue
				}
			}
			node = node.NextSibling
		}

		for child := parent.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				process(child)
			}
		}
	}

	process(root)
}

func combineAllInlineStyles(root *html.Node) {
	var styles []*html.Node
	var collect func(node *html.Node)
	collect = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "style" && isInlineStyle(node) {
			styles = append(styles, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			collect(child)
		}
	}

	collect(root)
	if len(styles) <= 1 {
		return
	}

	first := styles[0]
	parent := first.Parent
	if parent == nil {
		return
	}

	var importRules bytes.Buffer
	var otherRules bytes.Buffer
	for _, node := range styles {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != html.TextNode {
				continue
			}
			text := child.Data
			for _, line := range strings.Split(text, "\n") {
				trim := strings.TrimSpace(line)
				if trim == "" {
					continue
				}
				if strings.HasPrefix(trim, "@import") {
					importRules.WriteString(trim)
					if !strings.HasSuffix(trim, ";") {
						importRules.WriteString(";")
					}
					importRules.WriteString("\n")
				} else {
					otherRules.WriteString(line)
					otherRules.WriteString("\n")
				}
			}
		}
	}

	var combined bytes.Buffer
	combined.Write(importRules.Bytes())
	combined.Write(otherRules.Bytes())

	replacement := &html.Node{
		Type: html.ElementNode,
		Data: "style",
	}
	replacement.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: combined.String(),
	})

	parent.InsertBefore(replacement, first)
	for i := len(styles) - 1; i >= 0; i-- {
		styles[i].Parent.RemoveChild(styles[i])
	}
}

func combineGroup(parent *html.Node, tag string, group []*html.Node) {
	var combined bytes.Buffer
	for _, node := range group {
		if node.Type == html.ElementNode && node.Data == tag {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.TextNode {
					combined.WriteString(child.Data)
					combined.WriteString("\n")
				}
			}
		}
	}

	replacement := &html.Node{
		Type: html.ElementNode,
		Data: tag,
	}
	replacement.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: combined.String(),
	})

	first := group[0]
	parent.InsertBefore(replacement, first)
	for i := len(group) - 1; i >= 0; i-- {
		parent.RemoveChild(group[i])
	}
}

func walkHTMLNode(node *html.Node) {
	if node.Type == html.ElementNode {
		switch node.Data {
		case "script":
			if isInlineJavaScript(node) {
				minifyNodeContent(node, api.LoaderJS)
			}
		case "style":
			if isInlineStyle(node) {
				minifyNodeContent(node, api.LoaderCSS)
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkHTMLNode(child)
	}
}

func isInlineStyle(node *html.Node) bool {
	if hasAttr(node, "media") {
		return false
	}

	typeValue := strings.TrimSpace(strings.ToLower(getAttr(node, "type")))
	if typeValue == "" || typeValue == "text/css" {
		return true
	}

	return false
}

func isInlineJavaScript(node *html.Node) bool {
	if hasAttr(node, "src") {
		return false
	}

	typeValue := strings.TrimSpace(strings.ToLower(getAttr(node, "type")))
	if typeValue == "" {
		return true
	}

	if strings.Contains(typeValue, "javascript") || strings.Contains(typeValue, "module") {
		return true
	}

	return false
}

func hasAttr(node *html.Node, name string) bool {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return true
		}
	}
	return false
}

func getAttr(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func minifyNodeContent(node *html.Node, loader api.Loader) {
	var content bytes.Buffer
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			content.WriteString(child.Data)
		}
	}

	original := content.String()
	trimmed := strings.TrimSpace(original)
	if trimmed == "" {
		return
	}

	if loader == api.LoaderJS {
		original = minifyVueTemplateStrings(original)
	}

	result := api.Transform(original, api.TransformOptions{
		Loader:            loader,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		Sourcefile:        "inline",
	})

	if len(result.Errors) > 0 || len(result.Code) == 0 {
		return
	}

	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		node.RemoveChild(child)
		child = next
	}

	node.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: string(result.Code),
	})
}

var vueTemplateDirectiveRe = regexp.MustCompile(`\{\{\s*template\s+(['"])([^'\"]+)['"]\s*\}\}`)

func minifyVueTemplateStrings(source string) string {
	var out strings.Builder
	pos := 0
	for {
		idx := strings.Index(source[pos:], "template:")
		if idx < 0 {
			out.WriteString(source[pos:])
			break
		}
		idx += pos
		out.WriteString(source[pos:idx])
		j := idx + len("template:")
		for j < len(source) && (source[j] == ' ' || source[j] == '\t' || source[j] == '\n' || source[j] == '\r') {
			j++
		}
		if j >= len(source) {
			out.WriteString(source[idx:])
			break
		}
		delim := source[j]
		if delim != '`' && delim != '"' && delim != '\'' {
			out.WriteString(source[idx:j])
			pos = j
			continue
		}
		start := j
		end := start + 1
		for end < len(source) {
			if source[end] == '\\' {
				end += 2
				continue
			}
			if source[end] == delim {
				break
			}
			end++
		}
		if end >= len(source) {
			out.WriteString(source[idx:])
			break
		}

		out.WriteString(source[idx : start+1])
		content := source[start+1 : end]
		out.WriteString(minifyVueTemplateContent(content))
		out.WriteByte(delim)
		pos = end + 1
	}
	return out.String()
}

func minifyVueTemplateContent(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.Join(strings.Fields(value), " ")
	value = strings.ReplaceAll(value, "> <", "><")
	value = vueTemplateDirectiveRe.ReplaceAllString(value, `{{template $1$2$1}}`)
	return strings.TrimSpace(value)
}
