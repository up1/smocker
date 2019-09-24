package templates

import (
	"fmt"
	"net/http"

	"github.com/Thiht/smocker/types"
	json "github.com/layeh/gopher-json"
	log "github.com/sirupsen/logrus"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type luaEngine struct{}

func NewLuaEngine() TemplateEngine {
	return &luaEngine{}
}

func (*luaEngine) Execute(request types.Request, script string) *types.MockResponse {
	luaState := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})
	defer luaState.Close()

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
		{lua.MathLibName, lua.OpenMath},
		{lua.StringLibName, lua.OpenString},
		{lua.TabLibName, lua.OpenTable},
	} {
		if err := luaState.CallByParam(
			lua.P{
				Fn:      luaState.NewFunction(pair.f),
				NRet:    0,
				Protect: true,
			},
			lua.LString(pair.n),
		); err != nil {
			log.WithError(err).Error("Failed to load Lua libraries")
			return &types.MockResponse{
				Status: http.StatusInternalServerError,
				Body:   fmt.Sprintf("Failed to load Lua libraries: %s", err.Error()),
			}
		}
	}
	if err := luaState.DoString("coroutine=nil;debug=nil;io=nil;open=nil;os=nil"); err != nil {
		log.WithError(err).Error("Failed to sandbox Lua environment")
		return &types.MockResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Sprintf("Failed to sandbox Lua environment: %s", err.Error()),
		}
	}

	luaState.SetGlobal("request", luar.New(luaState, request))
	if err := luaState.DoString(script); err != nil {
		log.WithError(err).Error("Failed to execute dynamic template")
		return &types.MockResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Sprintf("Failed to execute dynamic template: %s", err.Error()),
		}
	}

	tmp := luaState.Get(-1).(*lua.LTable)
	body := tmp.RawGetString("body")
	if body.Type() == lua.LTTable {
		b, _ := json.Encode(body)
		tmp.RawSetString("body", lua.LString(string(b)))
	}

	var result types.MockResponse
	if err := gluamapper.Map(tmp, &result); err != nil {
		log.WithError(err).Error("Invalid result from Lua script")
		return &types.MockResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Sprintf("Invalid result from Lua script: %s", err.Error()),
		}
	}
	return &result
}