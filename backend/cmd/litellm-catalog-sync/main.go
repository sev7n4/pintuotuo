// Command litellm-catalog-sync: 以 DB 商品目录为单一事实来源，生成或校验 LiteLLM model_list 与 litellm_proxy_config.yaml 一致。
// 厂商→上游与密钥环境变量：见 deploy/litellm/provider_gateway_map.json 与 deploy/litellm/SSOT_ROUTING.md。
//
// 用法:
//
//	go run ./cmd/litellm-catalog-sync -verify -config ../../deploy/litellm/litellm_proxy_config.yaml -map ../../deploy/litellm/provider_gateway_map.json
//	go run ./cmd/litellm-catalog-sync -generate -map ../../deploy/litellm/provider_gateway_map.json -out generated.yaml
//
// 环境变量: DATABASE_URL（verify / generate 必填）
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
)

// providerMapEntry 对应 deploy/litellm/provider_gateway_map.json（单一配置入口）。
// litellm_model_template 优先：可含占位符 {model_id}；否则回退 litellm_prefix + "/" + model_id（兼容旧字段）。
type providerMapEntry struct {
	LitellmPrefix        string `json:"litellm_prefix,omitempty"`
	LitellmModelTemplate string `json:"litellm_model_template,omitempty"`
	APIKeyEnv            string `json:"api_key_env"`
	APIBase              string `json:"api_base,omitempty"`
	Notes                string `json:"notes,omitempty"`
}

func (e providerMapEntry) upstreamLitellmModel(modelID string) string {
	modelID = strings.TrimSpace(modelID)
	if t := strings.TrimSpace(e.LitellmModelTemplate); t != "" {
		return strings.ReplaceAll(t, "{model_id}", modelID)
	}
	if p := strings.TrimSpace(e.LitellmPrefix); p != "" {
		return p + "/" + modelID
	}
	return modelID
}

func main() {
	verify := flag.Bool("verify", false, "校验 yaml 中 model_name 覆盖库内 active SPU（需在 map 中的 provider）")
	generate := flag.Bool("generate", false, "从库生成 model_list 片段 YAML 到 -out")
	soft := flag.Bool("soft", false, "verify 时仅打印缺失项并以 0 退出（用于迁移期种子库与网关 P0 列表不一致）")
	configPath := flag.String("config", "", "litellm_proxy_config.yaml 路径")
	mapPath := flag.String("map", "", "provider_gateway_map.json 路径（平台厂商→LiteLLM 上游与 api_key 环境变量）")
	outPath := flag.String("out", "", "generate 输出文件路径（默认 stdout）")
	flag.Parse()

	if *verify && *generate {
		fmt.Fprintln(os.Stderr, "specify only one of -verify or -generate")
		os.Exit(2)
	}
	if !*verify && !*generate {
		fmt.Fprintln(os.Stderr, "specify -verify or -generate")
		os.Exit(2)
	}

	mapping, err := loadProviderMap(*mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map: %v\n", err)
		os.Exit(1)
	}

	if *verify {
		if *configPath == "" {
			fmt.Fprintln(os.Stderr, "-config required for -verify")
			os.Exit(2)
		}
		ok, err := runVerify(*configPath, mapping, *soft)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if ok {
			fmt.Println("verify: OK — 库内已映射厂商的 active SPU 均在 LiteLLM 配置中有对应 model_name")
		} else {
			fmt.Println("verify: soft mode — 已输出缺失项，未阻断（请逐步对齐 yaml 或目录）")
		}
		return
	}

	if err := runGenerate(mapping, *outPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func loadProviderMap(path string) (map[string]providerMapEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("-map required")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]providerMapEntry
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

var modelNameLine = regexp.MustCompile(`(?m)^\s*-\s*model_name:\s*(.+?)\s*$`)

func extractModelNamesFromYAML(content string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, m := range modelNameLine.FindAllStringSubmatch(content, -1) {
		name := strings.TrimSpace(m[1])
		name = strings.Trim(name, `"'`)
		out[name] = struct{}{}
	}
	return out
}

var fallbackBraceRE = regexp.MustCompile(`\{\s*"?([a-zA-Z0-9_.-]+)"?\s*:\s*\[([^\]]+)\]`)

func routerSettingsSection(content string) string {
	const start = "router_settings:"
	i := strings.Index(content, start)
	if i < 0 {
		return ""
	}
	rest := content[i:]
	j := strings.Index(rest, "\nlitellm_settings:")
	if j < 0 {
		return rest
	}
	return rest[:j]
}

// verifyFallbackModelNamesInList 检查 router_settings 内 fallbacks / context_window_fallbacks / content_policy_fallbacks
// 中出现的 model_name 是否均在 model_list 的 model_name 集合中（大小写不敏感）。
func stripCommentLines(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func verifyFallbackModelNamesInList(content string, modelNames map[string]struct{}) []string {
	sec := stripCommentLines(routerSettingsSection(content))
	if sec == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var missing []string
	for _, m := range fallbackBraceRE.FindAllStringSubmatch(sec, -1) {
		if len(m) < 3 {
			continue
		}
		addFallbackToken(strings.TrimSpace(m[1]), modelNames, seen, &missing)
		for _, part := range strings.Split(m[2], ",") {
			t := strings.Trim(strings.TrimSpace(part), `"'`)
			addFallbackToken(t, modelNames, seen, &missing)
		}
	}
	return missing
}

func addFallbackToken(token string, modelNames map[string]struct{}, seen map[string]struct{}, missing *[]string) {
	if token == "" {
		return
	}
	if _, ok := seen[token]; ok {
		return
	}
	seen[token] = struct{}{}
	if _, ok := modelNames[token]; ok {
		return
	}
	if nameSetContainsCI(modelNames, token) {
		return
	}
	*missing = append(*missing, token)
}

// runVerify 返回 ok=true 表示目录与 yaml 完全一致；soft 模式下有缺失时 ok=false 且 err=nil。
func runVerify(configPath string, mapping map[string]providerMapEntry, soft bool) (ok bool, err error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return false, fmt.Errorf("DATABASE_URL is required for -verify")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return false, err
	}
	defer db.Close()

	b, err := os.ReadFile(configPath)
	if err != nil {
		return false, err
	}
	content := string(b)
	names := extractModelNamesFromYAML(content)

	if fbMissing := verifyFallbackModelNamesInList(content, names); len(fbMissing) > 0 {
		msg := fmt.Sprintf("verify: router_settings 中 fallbacks 引用的 model_name 未在 model_list 中找到:\n  - %s",
			strings.Join(fbMissing, "\n  - "))
		if soft {
			fmt.Fprintln(os.Stderr, msg)
		} else {
			return false, fmt.Errorf("%s", msg)
		}
	}

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `
		SELECT mp.code,
			COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name) AS model_id
		FROM spus sp
		INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
		WHERE sp.status = 'active'
		ORDER BY mp.sort_order NULLS LAST, sp.sort_order NULLS LAST, sp.id
	`)
	if err != nil {
		return false, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	var missing []string
	for rows.Next() {
		var code, modelID string
		if err := rows.Scan(&code, &modelID); err != nil {
			return false, err
		}
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		if _, ok := mapping[code]; !ok {
			fmt.Fprintf(os.Stderr, "verify: skip provider %q (not in provider_gateway_map.json — 请补充映射或从校验范围排除)\n", code)
			continue
		}
		canonical := strings.ToLower(code + "/" + modelID)
		short := strings.ToLower(modelID)
		if _, ok := names[canonical]; ok {
			continue
		}
		if _, ok := names[short]; ok {
			continue
		}
		// 大小写不敏感再扫一遍
		if nameSetContainsCI(names, canonical) || nameSetContainsCI(names, short) {
			continue
		}
		missing = append(missing, canonical)
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	if len(missing) > 0 {
		msg := fmt.Sprintf("verify: 下列目录模型未在 %s 的 model_name 中找到（需短名或 canonical id）:\n  - %s",
			filepath.Base(configPath), strings.Join(missing, "\n  - "))
		if soft {
			fmt.Fprintln(os.Stderr, msg)
			return false, nil
		}
		return false, fmt.Errorf("%s", msg)
	}
	return true, nil
}

func nameSetContainsCI(names map[string]struct{}, want string) bool {
	w := strings.ToLower(want)
	for n := range names {
		if strings.ToLower(n) == w {
			return true
		}
	}
	return false
}

func runGenerate(mapping map[string]providerMapEntry, outPath string) error {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required for -generate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `
		SELECT mp.code,
			COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name) AS model_id
		FROM spus sp
		INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
		WHERE sp.status = 'active'
		ORDER BY mp.sort_order NULLS LAST, sp.sort_order NULLS LAST, sp.id
	`)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("# 由 litellm-catalog-sync -generate 生成；请勿手改本段（重新运行生成覆盖）\n")
	sb.WriteString("# 合并到 litellm_proxy_config.yaml 的 model_list 或用于与手工 P0 条目对照。\n\n")

	for rows.Next() {
		var code, modelID string
		if err := rows.Scan(&code, &modelID); err != nil {
			return err
		}
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		ent, ok := mapping[code]
		if !ok {
			fmt.Fprintf(os.Stderr, "generate: skip provider %q (not in map)\n", code)
			continue
		}
		litellmModel := ent.upstreamLitellmModel(modelID)
		shortName := modelID
		sb.WriteString(fmt.Sprintf("  - model_name: %s\n", shortName))
		sb.WriteString("    litellm_params:\n")
		sb.WriteString(fmt.Sprintf("      model: %s\n", litellmModel))
		sb.WriteString(fmt.Sprintf("      api_key: os.environ/%s\n", ent.APIKeyEnv))
		if b := strings.TrimSpace(ent.APIBase); b != "" {
			sb.WriteString(fmt.Sprintf("      api_base: %s\n", b))
		}
		sb.WriteString("\n")
	}
	if err := rows.Err(); err != nil {
		return err
	}

	out := sb.String()
	if outPath == "" {
		fmt.Print(out)
		return nil
	}
	return os.WriteFile(outPath, []byte(out), 0644)
}
