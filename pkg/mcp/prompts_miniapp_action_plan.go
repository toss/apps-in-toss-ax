package mcp

import (
	"context"
	_ "embed"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed prompts_miniapp_action_plan.md
var miniappActionPlanContent string

var miniappActionPlanCompletions = []Completion{
	{
		Ref:    PromptRef("miniapp-action-plan"),
		Arg:    "platform",
		Values: []string{"react-native", "web"},
	},
	{
		Ref:    PromptRef("miniapp-action-plan"),
		Arg:    "package_manager",
		Values: []string{"npm", "pnpm", "yarn"},
	},
}

var miniappActionPlan = &mcp.Prompt{
	Name:        "miniapp-action-plan",
	Title:       "AppsInToss Mini-App Development Action Plan",
	Description: "Provides a step-by-step action plan and checklist for developing an AppsInToss mini-app, including project initialization, framework setup, and launch preparation.",
	Arguments: []*mcp.PromptArgument{
		{
			Name:        "platform",
			Title:       "Platform",
			Description: "Development platform: react-native or web",
			Required:    true,
		},
		{
			Name:        "package_manager",
			Title:       "Package Manager",
			Description: "Package manager to use: npm, pnpm, or yarn",
			Required:    true,
		},
	},
}

func miniappActionPlanHandler(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	platform := req.Params.Arguments["platform"]
	packageManager := req.Params.Arguments["package_manager"]

	vars := buildPlatformVars(platform, packageManager)

	content := miniappActionPlanContent
	for k, v := range vars {
		content = strings.ReplaceAll(content, k, v)
	}

	return &mcp.GetPromptResult{
		Description: "AppsInToss mini-app development action plan for " + platform + " using " + packageManager,
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}

func buildPlatformVars(platform, packageManager string) map[string]string {
	vars := map[string]string{
		"{{platform}}":        platform,
		"{{package_manager}}": packageManager,
		"{{init_command}}":    resolveInitCommand(packageManager),
	}

	switch platform {
	case "web":
		vars["{{framework_package}}"] = "@apps-in-toss/web-framework"
		vars["{{tds_package}}"] = "@toss/tds-mobile"
		vars["{{routing_detail}}"] = "- Set up pages using file-based routing or manual route configuration"
		vars["{{platform_note}}"] = "- Ensure WebView compatibility with Toss in-app browser"
		vars["{{testing_checklist}}"] = "- [ ] 다양한 브라우저 환경에서 WebView 호환성 확인"
		vars["{{tds_tool_guide}}"] = "- `search_tds_web_docs`: TDS Web 컴포넌트 문서 검색\n- `get_tds_web_doc`: TDS Web 문서 상세 내용 조회"
	default: // react-native
		vars["{{framework_package}}"] = "@apps-in-toss/framework"
		vars["{{tds_package}}"] = "@toss/tds-react-native"
		vars["{{routing_detail}}"] = "- Set up screens using React Navigation or framework routing"
		vars["{{platform_note}}"] = "- Use React Native components for near-native performance"
		vars["{{testing_checklist}}"] = "- [ ] iOS / Android 기기에서 네이티브 동작 확인"
		vars["{{tds_tool_guide}}"] = "- `search_tds_rn_docs`: TDS React Native 컴포넌트 문서 검색\n- `get_tds_rn_doc`: TDS React Native 문서 상세 내용 조회"
	}

	return vars
}

func resolveInitCommand(packageManager string) string {
	switch packageManager {
	case "pnpm":
		return "pnpm dlx ait init"
	case "yarn":
		return "yarn dlx ait init"
	default:
		return "npx ait init"
	}
}
