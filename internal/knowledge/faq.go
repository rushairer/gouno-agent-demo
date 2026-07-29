package knowledge

import (
	"strings"
	"unicode"
)

// Article is deliberately small and local: the demo never lets the model invent a knowledge source.
type Article struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Keywords []string `json:"-"`
	Content  string   `json:"content"`
}

var articles = []Article{
	{ID: "vpn", Title: "VPN 连接失败", Keywords: []string{"vpn", "内网", "连接", "远程", "network", "remote"}, Content: "确认网络可用后，重新打开 VPN 客户端并使用公司账号登录。若仍失败，请记录错误码并提交网络服务工单。"},
	{ID: "password", Title: "密码重置", Keywords: []string{"密码", "重置", "忘记", "账号", "password", "account"}, Content: "请使用公司身份门户的“忘记密码”流程完成验证和重置。不要向任何人提供验证码或密码。"},
	{ID: "mfa", Title: "MFA 多因素认证", Keywords: []string{"mfa", "双因素", "验证码", "认证器", "two-factor", "authenticator"}, Content: "更换设备或无法取得验证码时，请通过身份门户的恢复流程验证身份，再重新绑定认证器。"},
	{ID: "printer", Title: "打印机故障", Keywords: []string{"打印机", "打印", "纸张", "驱动", "printer", "print", "driver"}, Content: "检查设备在线、纸张和队列状态；清除卡住的任务并重试。持续故障请附设备编号提交桌面支持工单。"},
	{ID: "ticket", Title: "工单升级", Keywords: []string{"工单", "升级", "紧急", "服务台", "ticket", "urgent", "service desk"}, Content: "工单中应提供影响范围、开始时间、错误截图和已尝试步骤。影响多人或核心业务时选择紧急等级并联系值班服务台。"},
}

func Search(input string) (Article, bool) {
	input = strings.ToLower(input)
	var best Article
	bestScore := 0
	for _, article := range articles {
		score := 0
		for _, keyword := range article.Keywords {
			if strings.Contains(input, keyword) {
				score++
			}
		}
		if score > bestScore {
			best, bestScore = article, score
		}
	}
	return best, bestScore > 0
}

func Normalise(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, strings.TrimSpace(input))
}
