package setting

import (
	"github.com/QuantumNous/new-api/common"
)

// CustomMenuItem 自定义侧边栏菜单项。
// 一条记录 = 侧边栏「聊天」分组下方的一个菜单项。
//   - Name: 显示文字
//   - URL : 跳转地址(支持 {address} / {key} 占位符,前端 SiderBar 渲染时替换)
//   - Icon: lucide-react 图标名(大小写敏感,例如 "Bot"、"Sparkles"、"Wand2");
//     缺省或解析失败时会回退到默认图标。
type CustomMenuItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}

// CustomMenu 全局可变,通过 options 表的 "CustomMenu" key 持久化。
var CustomMenu = []CustomMenuItem{}

func UpdateCustomMenuByJsonString(jsonString string) error {
	CustomMenu = make([]CustomMenuItem, 0)
	if jsonString == "" {
		return nil
	}
	return common.UnmarshalJsonStr(jsonString, &CustomMenu)
}

func CustomMenu2JsonString() string {
	if len(CustomMenu) == 0 {
		return "[]"
	}
	jsonBytes, err := common.Marshal(CustomMenu)
	if err != nil {
		common.SysLog("error marshalling custom menu: " + err.Error())
		return "[]"
	}
	return string(jsonBytes)
}
