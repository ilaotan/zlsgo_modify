package cors

import (
	"net/http"
	"strings"

	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Config cors configuration
	Config struct {
		CustomHandler Handler
		methods       string
		credentials   string
		headers       string
		exposeHeaders string
		Domains       []string
		Methods       []string
		Credentials   []string
		Headers       []string
		ExposeHeaders []string
	}
	Handler func(conf *Config, c *znet.Context)
)

func Default() znet.HandlerFunc {
	return New(&Config{})
}

func NewAllowHeaders() (addAllowHeader func(header string), handler znet.HandlerFunc) {
	conf := &Config{}
	handler = New(conf)

	return func(header string) {
		headers := strings.Split(conf.headers, ", ")
		headers = append(headers, header)
		conf.headers = strings.Join(headers, ", ")
	}, handler
}

func New(conf *Config) znet.HandlerFunc {
	if len(conf.Methods) == 0 {
		conf.Methods = []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}
	}
	conf.methods = strings.Join(conf.Methods, ", ")
	if len(conf.Credentials) == 0 {
		conf.Credentials = []string{"true"}
	}
	conf.credentials = strings.Join(conf.Credentials, ", ")
	if len(conf.Headers) == 0 {
		conf.Headers = []string{"Origin", "No-Cache", "X-Requested-With", "If-Modified-Since", "Pragma", "Last-Modified", "Cache-Control", "Expires", "Content-Type", "Access-Control-Allow-Origin", "Authorization"}
	}
	conf.headers = strings.Join(conf.Headers, ", ")

	if len(conf.ExposeHeaders) > 0 {
		conf.exposeHeaders = strings.Join(conf.ExposeHeaders, ", ")
	}

	return func(c *znet.Context) {
		if applyCors(c, conf) {
			c.Next()
		}
	}
}

func applyCors(c *znet.Context, conf *Config) bool {
	origin := c.GetHeader("Origin")
	if len(origin) == 0 {
		return true
	}

	domains := conf.Domains
	if len(domains) > 0 {
		adopt := false
		for k := range domains {
			if adopt = zstring.Match(origin, domains[k]); adopt {
				break
			}
		}
		if !adopt {
			c.Abort(http.StatusForbidden)
			return false
		}
	}

	c.SetHeader("Access-Control-Allow-Methods", conf.methods)
	c.SetHeader("Access-Control-Allow-Credentials", conf.credentials)
	c.SetHeader("Access-Control-Allow-Headers", conf.headers)
	if conf.exposeHeaders != "" {
		c.SetHeader("Access-Control-Expose-Headers", conf.exposeHeaders)
	}
	c.SetHeader("Access-Control-Allow-Origin", origin)
	if conf.CustomHandler != nil {
		conf.CustomHandler(conf, c)
	}

	if c.Request.Method == "OPTIONS" {
		c.Abort(http.StatusNoContent)
		return false
	}

	return true
}
