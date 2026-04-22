package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// ModelFallbackRule 对应 model_fallback_rules 表行（管理端 CRUD）。
type ModelFallbackRule struct {
	ID             int      `json:"id"`
	SourceModel    string   `json:"source_model"`
	FallbackModels []string `json:"fallback_models"`
	Enabled        bool     `json:"enabled"`
	Notes          string   `json:"notes,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
}

// CanonicalCatalogModelKey 将输入规范为目录中真实存在的 `code/模型标识`（与 GET /openai/v1/models 的 id 一致）。
func CanonicalCatalogModelKey(ctx context.Context, db *sql.DB, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("模型键不能为空")
	}
	idx := strings.Index(raw, "/")
	if idx <= 0 || idx == len(raw)-1 {
		return "", fmt.Errorf("模型键须为 provider/model 形式")
	}
	prov := strings.ToLower(strings.TrimSpace(raw[:idx]))
	name := strings.TrimSpace(raw[idx+1:])
	if prov == "" || name == "" {
		return "", fmt.Errorf("provider 或模型名为空")
	}
	var code, mid string
	err := db.QueryRowContext(ctx, `
		SELECT mp.code, TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name))
		FROM spus sp
		INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
		WHERE sp.status = 'active' AND LOWER(mp.code) = $1
		  AND LOWER(TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name))) = LOWER($2)
		ORDER BY sp.id ASC
		LIMIT 1
	`, prov, name).Scan(&code, &mid)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("模型不在上架目录中: %s/%s", prov, name)
	}
	if err != nil {
		return "", err
	}
	mid = strings.TrimSpace(mid)
	if mid == "" {
		return "", fmt.Errorf("模型不在上架目录中: %s/%s", prov, name)
	}
	return code + "/" + mid, nil
}

// DedupeFallbackChainUnique 保留顺序、去掉重复项（去重）。
func DedupeFallbackChainUnique(canonicals []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, c := range canonicals {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

// FallbackGraphHasCycle 检测有向图中是否存在环（任意环均视为非法 fallback 配置）。
func FallbackGraphHasCycle(g map[string][]string) bool {
	state := make(map[string]int8) // 0=unseen, 1=stack, 2=done
	var visit func(string) bool
	visit = func(u string) bool {
		switch state[u] {
		case 1:
			return true
		case 2:
			return false
		}
		state[u] = 1
		for _, v := range g[u] {
			if visit(v) {
				return true
			}
		}
		state[u] = 2
		return false
	}
	for n := range g {
		if state[n] == 0 && visit(n) {
			return true
		}
	}
	return false
}

// LoadEnabledFallbackGraph 从 DB 加载启用规则的图；excludeID>0 时跳过该行（用于更新前替换）。
func LoadEnabledFallbackGraph(ctx context.Context, db *sql.DB, excludeID int) (map[string][]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, source_model, fallback_models FROM model_fallback_rules WHERE enabled = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	g := make(map[string][]string)
	for rows.Next() {
		var id int
		var src string
		var fbs pq.StringArray
		if err := rows.Scan(&id, &src, &fbs); err != nil {
			return nil, err
		}
		if excludeID > 0 && id == excludeID {
			continue
		}
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}
		list := make([]string, 0, len(fbs))
		for _, x := range fbs {
			x = strings.TrimSpace(x)
			if x != "" {
				list = append(list, x)
			}
		}
		g[src] = DedupeFallbackChainUnique(list)
	}
	return g, rows.Err()
}

// ValidateFallbackRule 校验并规范化 source 与 fallback 链；合并图中边后检测环。
func ValidateFallbackRule(ctx context.Context, db *sql.DB, excludeID int, sourceRaw string, fallbackRaws []string, enabled bool) (source string, fallbacks []string, err error) {
	source, err = CanonicalCatalogModelKey(ctx, db, sourceRaw)
	if err != nil {
		return "", nil, err
	}
	var chain []string
	for i, raw := range fallbackRaws {
		c, e := CanonicalCatalogModelKey(ctx, db, raw)
		if e != nil {
			return "", nil, fmt.Errorf("备用[%d]: %w", i, e)
		}
		chain = append(chain, c)
	}
	chain = DedupeFallbackChainUnique(chain)
	if enabled && len(chain) == 0 {
		return "", nil, fmt.Errorf("启用规则时至少需要一个备用模型")
	}
	for _, f := range chain {
		if f == source {
			return "", nil, fmt.Errorf("备用链不能包含与主模型相同的条目（自 fallback）")
		}
	}

	g, err := LoadEnabledFallbackGraph(ctx, db, excludeID)
	if err != nil {
		return "", nil, err
	}
	if g == nil {
		g = make(map[string][]string)
	}
	if enabled {
		g[source] = chain
	}
	if FallbackGraphHasCycle(g) {
		return "", nil, fmt.Errorf("fallback 配置会形成循环依赖，请调整链或禁用冲突规则")
	}
	return source, chain, nil
}

// GetEnabledFallbackChain 返回主模型（规范化后的 catalog key）对应的启用备用链，无规则时 nil。
func GetEnabledFallbackChain(ctx context.Context, db *sql.DB, sourceCanonical string) ([]string, error) {
	if db == nil || strings.TrimSpace(sourceCanonical) == "" {
		return nil, nil
	}
	var fbs pq.StringArray
	err := db.QueryRowContext(ctx,
		`SELECT fallback_models FROM model_fallback_rules WHERE enabled = true AND source_model = $1`,
		strings.TrimSpace(sourceCanonical),
	).Scan(&fbs)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, x := range fbs {
		x = strings.TrimSpace(x)
		if x != "" {
			out = append(out, x)
		}
	}
	return DedupeFallbackChainUnique(out), nil
}

// SplitCatalogModelKey 解析目录键 provider/model（小写 provider code）。
func SplitCatalogModelKey(key string) (provider string, modelName string, err error) {
	key = strings.TrimSpace(key)
	idx := strings.Index(key, "/")
	if idx <= 0 || idx == len(key)-1 {
		return "", "", fmt.Errorf("invalid catalog model key")
	}
	p := strings.ToLower(strings.TrimSpace(key[:idx]))
	m := strings.TrimSpace(key[idx+1:])
	if p == "" || m == "" {
		return "", "", fmt.Errorf("invalid catalog model key")
	}
	return p, m, nil
}

// ListCatalogModelKeys 返回当前上架目录中全部 OpenAI 兼容模型 id（provider/model）。
func ListCatalogModelKeys(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT mp.code || '/' || TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name))
		FROM spus sp
		INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
		WHERE sp.status = 'active'
		ORDER BY mp.sort_order NULLS LAST, sp.sort_order NULLS LAST, sp.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		k = strings.TrimSpace(k)
		if k != "" && strings.Contains(k, "/") {
			keys = append(keys, k)
		}
	}
	return keys, rows.Err()
}

// ListModelFallbackRules 列出全部规则（管理端）。
func ListModelFallbackRules(ctx context.Context, db *sql.DB) ([]ModelFallbackRule, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_model, fallback_models, enabled, COALESCE(notes,''),
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM model_fallback_rules
		ORDER BY source_model ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ModelFallbackRule
	for rows.Next() {
		var r ModelFallbackRule
		var fbs pq.StringArray
		if err := rows.Scan(&r.ID, &r.SourceModel, &fbs, &r.Enabled, &r.Notes, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		for _, x := range fbs {
			x = strings.TrimSpace(x)
			if x != "" {
				r.FallbackModels = append(r.FallbackModels, x)
			}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
