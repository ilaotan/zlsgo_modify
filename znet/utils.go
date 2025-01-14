package znet

import (
	"errors"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

func getAddr(addr string) string {
	var port int
	if strings.Contains(addr, ":") {
		port, _ = strconv.Atoi(strings.Split(addr, ":")[1])
	} else {
		port, _ = strconv.Atoi(addr)
		addr = ":" + addr
	}
	if port != 0 {
		return addr
	}
	port, _ = Port(port, true)
	return ":" + strconv.Itoa(port)
}

func getHostname(addr string, isTls bool) string {
	hostname := "http://"
	if isTls {
		hostname = "https://"
	}
	return hostname + resolveHostname(addr)
}

func CompletionPath(p, prefix string) string {
	if p == "/" {
		if prefix == "/" {
			return prefix
		}
		return prefix + p
	} else if p == "" {
		if prefix[0] != '/' {
			return "/" + prefix
		}
		return prefix
	}
	suffix := false
	if p[len(p)-1] == '/' {
		suffix = true
	}
	p = zstring.TrimSpace(path.Join("/", prefix, p))
	if suffix {
		p = p + "/"
	}
	return p
}

func resolveAddr(addrString string, tlsConfig ...TlsCfg) addrSt {
	cfg := addrSt{
		addr: addrString,
	}
	if len(tlsConfig) > 0 {
		cfg.Cert = tlsConfig[0].Cert
		cfg.HTTPAddr = tlsConfig[0].HTTPAddr
		cfg.HTTPProcessing = tlsConfig[0].HTTPProcessing
		cfg.Key = tlsConfig[0].Key
		cfg.Config = tlsConfig[0].Config
	}
	return cfg
}

func resolveHostname(addrString string) string {
	if strings.Index(addrString, ":") == 0 {
		return "127.0.0.1" + addrString
	}
	return addrString
}

func templateParse(templateFile []string, funcMap template.FuncMap) (t *template.Template, err error) {
	if len(templateFile) == 0 {
		return nil, errors.New("template file cannot be empty")
	}
	file := templateFile[0]
	if len(file) <= 255 && zfile.FileExist(file) {
		for i := range templateFile {
			templateFile[i] = zfile.RealPath(templateFile[i])
		}
		t, err = template.ParseFiles(templateFile...)
		if err == nil && funcMap != nil {
			t.Funcs(funcMap)
		}
	} else {
		t = template.New("")
		if funcMap != nil {
			t.Funcs(funcMap)
		}
		t, err = t.Parse(file)
	}
	return
}

func parsPattern(res []string, prefix string) (string, []string) {
	var (
		matchName []string
		pattern   string
	)
	for _, str := range res {
		if str == "" {
			continue
		}
		pattern = pattern + prefix
		l := len(str) - 1
		i := strings.Index(str, "}")
		i2 := strings.Index(str, "{")
		firstChar := string(str[0])
		// todo Need to optimize
		if i2 != -1 && i != -1 {
			// lastChar := string(str[mu])
			if i == l && i2 == 0 {
				matchStr := str[1:l]
				res := strings.Split(matchStr, ":")
				matchName = append(matchName, res[0])
				pattern = pattern + "(" + res[1] + ")"
			} else {
				if i2 != 0 {
					p, m := parsPattern([]string{str[:i2]}, "")
					if p != "" {
						pattern = pattern + p
						matchName = append(matchName, m...)
					}
					str = str[i2:]
				}
				if i >= 0 {
					ni := i - i2
					matchStr := str[1:ni]
					res := strings.Split(matchStr, ":")
					matchName = append(matchName, res[0])
					pattern = pattern + "(" + res[1] + ")"
					p, m := parsPattern([]string{str[ni+1:]}, "")
					if p != "" {
						pattern = pattern + p
						matchName = append(matchName, m...)
					}
				} else {
					pattern = pattern + str
				}
			}

		} else if firstChar == ":" {
			matchStr := str
			res := strings.Split(matchStr, ":")
			key := res[1]
			if key == "full" {
				key = allKey
			}
			matchName = append(matchName, key)
			if key == idKey {
				pattern = pattern + "(" + idPattern + ")"
			} else if key == allKey {
				pattern = pattern + "(" + allPattern + ")"
			} else {
				pattern = pattern + "(" + defaultPattern + ")"
			}
		} else if firstChar == "*" {
			pattern = pattern + "(" + allPattern + ")"
			matchName = append(matchName, allKey)
		} else {
			pattern = pattern + str
		}
	}
	return pattern, matchName
}

type tlsRedirectHandler struct {
	Domain string
}

func (t *tlsRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, t.Domain+r.URL.String(), http.StatusMovedPermanently)
}

func (e *Engine) NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:        w,
		Request:       req,
		Engine:        e,
		Log:           e.Log,
		Cache:         Cache,
		startTime:     time.Time{},
		header:        map[string][]string{},
		customizeData: map[string]interface{}{},
		stopHandle:    zutil.NewBool(false),
		done:          zutil.NewBool(false),
		prevData: &PrevData{
			Code: zutil.NewInt32(0),
			Type: ContentTypePlain,
		},
	}
}

func (c *Context) clone(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Writer = w
	c.injector = zdi.New(c.Engine.injector)
	c.injector.Maps(c)
	c.startTime = time.Now()
	c.renderError = defErrorHandler()
	c.stopHandle.Store(false)
	c.done.Store(false)
}

func (e *Engine) acquireContext() *Context {
	return e.pool.Get().(*Context)
}

func (e *Engine) releaseContext(c *Context) {
	c.prevData.Code.Store(0)
	c.mu.Lock()
	c.middleware = c.middleware[0:0]
	c.customizeData = map[string]interface{}{}
	c.header = map[string][]string{}
	c.render = nil
	c.renderError = nil
	c.cacheJSON = nil
	c.cacheQuery = nil
	c.cacheForm = nil
	c.injector = nil
	c.rawData = c.rawData[0:0]
	c.prevData.Content = c.prevData.Content[0:0]
	c.prevData.Type = ContentTypePlain
	c.mu.Unlock()
	e.pool.Put(c)
}
