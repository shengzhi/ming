// 路由模块

package ming

import (
	"fmt"
	"reflect"
	"sort"
)

import "strings"

// ServiceName 服务名称类型
type ServiceName string

type routeGroup struct {
	routes  []*Route
	funcMap map[string]handlerFuncChain
}

func newRouteGroup() *routeGroup {
	return &routeGroup{}
}

func (rg *routeGroup) getHandler(h APIHeader) (handlerFuncChain, bool) {
	key := strings.ToLower(h.String())
	chain, has := rg.funcMap[key]
	return chain, has
}

func (rg *routeGroup) add(r *Route) {
	rg.routes = append(rg.routes, r)
}

func (rg *routeGroup) parseActions() {
	rg.funcMap = make(map[string]handlerFuncChain)
	for _, r := range rg.routes {
		for _, ctrl := range r.handlers {
			for _, name := range r.parseAction(ctrl) {
				var chain handlerFuncChain
				chain = append(chain, r.chain...)
				key := fmt.Sprintf("%s_%s_%s_%s_%s", r.version, r.service, r.module, name.controller, name.action)
				key = strings.ToLower(key)
				if _, ok := rg.funcMap[key]; ok {
					panic(fmt.Errorf("duplicated service %s", key))
				} else {
					chain = append(chain, name.fn)
				}
				rg.funcMap[key] = chain
			}
		}
	}
}

// actions 返回所有Action
func (rg routeGroup) actions() string {
	actions := make([]string, 0, len(rg.funcMap))
	for k := range rg.funcMap {
		actions = append(actions, k)
	}
	sort.Strings(actions)
	return strings.Join(actions, "\r\n")
}

// Route 路由
type Route struct {
	version, service, module string
	chain                    handlerFuncChain
	handlers                 []interface{}
}

// Register 注册处理器
func (r *Route) Register(controllers ...interface{}) {
	r.handlers = append(r.handlers, controllers...)
}

type funcName struct {
	controller string
	action     string
	fn         HandlerFunc
}

var handlerType = reflect.TypeOf(new(HandlerFunc)).Elem()
var seriveNameType = reflect.TypeOf(new(ServiceName)).Elem()

func (r *Route) parseAction(controller interface{}) (funcs []funcName) {
	v := reflect.ValueOf(controller)
	t := v.Type()
	st := t
	if t.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	ctrlName := st.Name()
	for i := 0; i < st.NumField(); i++ {
		if st.Field(i).Type == seriveNameType {
			if name, ok := st.Field(i).Tag.Lookup("api"); ok {
				ctrlName = name
			}
			break
		}
	}
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Method(i)
		if !method.Type().ConvertibleTo(handlerType) {
			continue
		}
		if fn, ok := method.Convert(handlerType).Interface().(HandlerFunc); ok {
			funcs = append(funcs, funcName{
				controller: ctrlName,
				action:     t.Method(i).Name,
				fn:         fn,
			})
		}
	}
	return
}
