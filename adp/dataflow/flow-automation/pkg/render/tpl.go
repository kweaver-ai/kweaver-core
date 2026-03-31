package render

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"text/template"

	"github.com/golang/groupcache/lru"
)

type TplProvider struct {
	cache   *lru.Cache
	rwMutex sync.RWMutex
}

func NewCachedTplProvider(maxSize int) *TplProvider {
	cache := lru.New(maxSize)
	return &TplProvider{
		cache:   cache,
		rwMutex: sync.RWMutex{},
	}
}

func (c *TplProvider) cacheGetTpl(tplText string) (*template.Template, bool) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	v, ok := c.cache.Get(tplText)
	if !ok {
		return nil, false
	}
	return v.(*template.Template), true
}

func (c *TplProvider) cacheSetTpl(tplText string, template *template.Template) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	c.cache.Add(tplText, template)
}

func (c *TplProvider) GetTpl(tplText string) (*template.Template, error) {
	tpl, ok := c.cacheGetTpl(tplText)
	if ok {
		return tpl, nil
	}
	tpl, err := c.parseTpl(tplText)
	if err != nil {
		return nil, err
	}
	c.cacheSetTpl(tplText, tpl)
	return tpl, err
}

func (c *TplProvider) parseTpl(tplText string) (*template.Template, error) {
	tplText = strings.ReplaceAll(tplText, "{{", "{{renderField ") 
	funcMap := template.FuncMap{
		"renderField": func(v interface{}) (interface{}, error) {
			val := reflect.ValueOf(v)

			switch val.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return string(b), nil
			default:
				return v, nil
			}
		},
	}

	tpl, err := template.New(tplText).Funcs(funcMap).Parse(tplText)
	if err != nil {
		return nil, err
	}
	tpl.Option("missingkey=error")
	return tpl, err
}
