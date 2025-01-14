package znet

import (
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	"github.com/sohaha/zlsgo/zdi"
)

// Host Get the current Host
func (c *Context) Host(full ...bool) string {
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "https"
		if c.Request.TLS == nil {
			scheme = "http"
		}
	}
	host := c.Request.Host
	if len(full) > 0 && full[0] {
		host += c.Request.URL.String()
	}
	return scheme + "://" + host
}

// CompletionLink Complete the link and add the current domain name if it is not linked
func (c *Context) CompletionLink(link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}
	finalLink := c.Host()
	if !strings.HasPrefix(link, "/") {
		finalLink = finalLink + "/"
	}
	finalLink = finalLink + link
	return finalLink
}

// IsWebsocket Is Websocket
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") &&
		strings.ToLower(c.GetHeader("Upgrade")) == "websocket" {
		return true
	}
	return false
}

// IsSSE Is SSE
func (c *Context) IsSSE() bool {
	return strings.ToLower(c.GetHeader("Accept")) == "text/event-stream"
}

// IsAjax IsAjax
func (c *Context) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// GetClientIP Client IP
func (c *Context) GetClientIP() (IP string) {
	IP = ClientPublicIP(c.Request)
	if IP == "" {
		IP = ClientIP(c.Request)
	}
	return
}

// GetHeader Get Header
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader Set Header
func (c *Context) SetHeader(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	c.mu.Lock()
	if value == "" {
		delete(c.header, key)
	} else {
		c.header[key] = append(c.header[key], value)
	}
	c.mu.Unlock()
}

func (c *Context) write() {
	if !c.done.CAS(false, true) {
		return
	}

	c.Next()

	data := c.PrevContent()
	// data.Code.CAS(0, http.StatusInternalServerError)

	for key, value := range c.header {
		for i := range value {
			header := value[i]
			if i == 0 {
				c.Writer.Header().Set(key, header)
			} else {
				c.Writer.Header().Add(key, header)
			}
		}
	}

	if c.Request == nil || c.Request.Context().Err() != nil {
		return
	}

	defer func() {
		if c.Engine.IsDebug() {
			requestLog(c)
		}
	}()

	code := int(data.Code.Load())
	if code == 0 {
		code = http.StatusOK
		data.Code.Store(int32(code))
	}
	size := len(data.Content)
	if size > 0 {
		c.Writer.Header().Set("Content-Length", strconv.Itoa(size))
		c.Writer.WriteHeader(code)
		_, err := c.Writer.Write(data.Content)
		if err != nil {
			c.Log.Error(err)
		}
		return
	}
	if code != 200 {
		c.Writer.WriteHeader(code)
	}
}

// Next Handler
func (c *Context) Next() {
	for {
		if c.stopHandle.Load() {
			break
		}
		c.mu.RLock()
		n := len(c.middleware) > 0
		c.mu.RUnlock()
		if !n {
			break
		}
		c.next()
	}
}

func (c *Context) next() {
	if c.stopHandle.Load() {
		return
	}
	c.mu.Lock()
	n := c.middleware[0]
	c.middleware = c.middleware[1:]
	c.mu.Unlock()
	err := n(c)
	if err != nil {
		c.Abort()
		c.renderError(c, err)
	}
}

// SetCookie Set Cookie
func (c *Context) SetCookie(name, value string, maxAge ...int) {
	a := 0
	if len(maxAge) > 0 {
		a = maxAge[0]
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   a,
	}
	c.Writer.Header().Add("Set-Cookie", cookie.String())
}

// GetCookie Get Cookie
func (c *Context) GetCookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	v, _ := url.QueryUnescape(cookie.Value)
	return v
}

// GetReferer request referer
func (c *Context) GetReferer() string {
	return c.Request.Header.Get("Referer")
}

// GetUserAgent http request UserAgent
func (c *Context) GetUserAgent() string {
	return c.Request.Header.Get("User-Agent")
}

// ContentType returns the Content-Type header of the request
func (c *Context) ContentType(contentText ...string) string {
	var content string
	if len(contentText) > 0 {
		content = contentText[0]
	} else {
		content = c.GetHeader("Content-Type")
	}
	for i := 0; i < len(content); i++ {
		char := content[i]
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

// WithValue context sharing data
func (c *Context) WithValue(key string, value interface{}) *Context {
	c.mu.Lock()
	c.customizeData[key] = value
	c.mu.Unlock()
	return c
}

// Value get context sharing data
func (c *Context) Value(key string, def ...interface{}) (value interface{}, ok bool) {
	c.mu.RLock()
	value, ok = c.customizeData[key]
	if !ok && (len(def) > 0) {
		value = def[0]
	}
	c.mu.RUnlock()
	return
}

// Value get context sharing data
func (c *Context) MustValue(key string, def ...interface{}) (value interface{}) {
	value, _ = c.Value(key, def)
	return
}

func (c *Context) Injector() zdi.Injector {
	return c.injector
}
