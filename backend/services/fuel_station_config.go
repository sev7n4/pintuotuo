package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pintuotuo/backend/config"
)

const (
	PlatformKeyFuelStationConfig          = "fuel_station_config"
	PlatformKeyFuelStationTemplateLibrary = "fuel_station_template_library"
)

type FuelStationTierConfig struct {
	Label string `json:"label"`
	SKUID int    `json:"sku_id"`
}

type FuelStationSectionConfig struct {
	Code        string                `json:"code"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Badge       string                `json:"badge"`
	SortOrder   int                   `json:"sort_order"`
	Status      string                `json:"status"`
	Tiers       []FuelStationTierConfig `json:"tiers"`
}

type FuelStationConfig struct {
	PageTitle    string                    `json:"page_title"`
	PageSubtitle string                    `json:"page_subtitle"`
	RuleText     string                    `json:"rule_text"`
	Sections     []FuelStationSectionConfig `json:"sections"`
}

type FuelStationTemplate struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Payload     FuelStationConfig `json:"payload"`
}

func defaultFuelStationConfig() FuelStationConfig {
	return FuelStationConfig{
		PageTitle:    "智燃加油站",
		PageSubtitle: "面向已订购模型权益用户，按用途与模型覆盖范围补充 Token。",
		RuleText:     "加油包不可单独购买，需与至少一个在售模型商品或套餐包组合下单。",
		Sections: []FuelStationSectionConfig{
			{
				Code:        "coding",
				Name:        "编程加油包",
				Description: "覆盖代码生成、代码审查、文档生成等编程场景。",
				Badge:       "编程主推",
				SortOrder:   10,
				Status:      "active",
				Tiers: []FuelStationTierConfig{
					{Label: "S 档", SKUID: 0},
					{Label: "M 档", SKUID: 0},
					{Label: "L 档", SKUID: 0},
				},
			},
		},
	}
}

func defaultFuelStationTemplateLibrary() []FuelStationTemplate {
	return []FuelStationTemplate{
		{
			Key:         "coding_v1",
			Name:        "编程品类 S/M/L（骨架）",
			Description: "默认一张编程卡片，S/M/L 三档 SKU 预留，便于快速开工。",
			Payload:     defaultFuelStationConfig(),
		},
		{
			Key:         "multi_scene_v1",
			Name:        "多场景扩展（骨架）",
			Description: "预留编程 + 音视频 + 绘画三卡片结构，便于后续扩展。",
			Payload: FuelStationConfig{
				PageTitle:    "智燃加油站",
				PageSubtitle: "按已订购模型类型，选择对应场景加油包。",
				RuleText:     "加油包不可单独购买，需与至少一个在售模型商品或套餐包组合下单。",
				Sections: []FuelStationSectionConfig{
					{
						Code:        "coding",
						Name:        "编程加油包",
						Description: "编程场景消耗补给。",
						Badge:       "编程主推",
						SortOrder:   10,
						Status:      "active",
						Tiers: []FuelStationTierConfig{
							{Label: "S 档", SKUID: 0},
							{Label: "M 档", SKUID: 0},
							{Label: "L 档", SKUID: 0},
						},
					},
					{
						Code:        "audio_video",
						Name:        "音视频加油包",
						Description: "音视频理解与生成场景补给。",
						Badge:       "待扩展",
						SortOrder:   20,
						Status:      "inactive",
						Tiers: []FuelStationTierConfig{
							{Label: "S 档", SKUID: 0},
							{Label: "M 档", SKUID: 0},
							{Label: "L 档", SKUID: 0},
						},
					},
					{
						Code:        "image",
						Name:        "绘画加油包",
						Description: "图像生成与编辑场景补给。",
						Badge:       "待扩展",
						SortOrder:   30,
						Status:      "inactive",
						Tiers: []FuelStationTierConfig{
							{Label: "S 档", SKUID: 0},
							{Label: "M 档", SKUID: 0},
							{Label: "L 档", SKUID: 0},
						},
					},
				},
			},
		},
	}
}

func normalizeFuelStationConfig(cfg FuelStationConfig) FuelStationConfig {
	if cfg.PageTitle == "" {
		cfg.PageTitle = defaultFuelStationConfig().PageTitle
	}
	if cfg.PageSubtitle == "" {
		cfg.PageSubtitle = defaultFuelStationConfig().PageSubtitle
	}
	if cfg.RuleText == "" {
		cfg.RuleText = defaultFuelStationConfig().RuleText
	}
	if len(cfg.Sections) == 0 {
		cfg.Sections = defaultFuelStationConfig().Sections
	}
	for i := range cfg.Sections {
		if cfg.Sections[i].Status == "" {
			cfg.Sections[i].Status = "active"
		}
	}
	return cfg
}

func GetFuelStationConfig(ctx context.Context) (FuelStationConfig, error) {
	db := config.GetDB()
	if db == nil {
		return FuelStationConfig{}, fmt.Errorf("database not initialized")
	}
	var raw string
	err := db.QueryRowContext(ctx, `SELECT value FROM platform_settings WHERE key = $1`, PlatformKeyFuelStationConfig).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultFuelStationConfig(), nil
	}
	if err != nil {
		return FuelStationConfig{}, err
	}
	var cfg FuelStationConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return defaultFuelStationConfig(), nil
	}
	return normalizeFuelStationConfig(cfg), nil
}

func UpdateFuelStationConfig(ctx context.Context, cfg FuelStationConfig) (FuelStationConfig, error) {
	db := config.GetDB()
	if db == nil {
		return FuelStationConfig{}, fmt.Errorf("database not initialized")
	}
	cfg = normalizeFuelStationConfig(cfg)
	data, err := json.Marshal(cfg)
	if err != nil {
		return FuelStationConfig{}, err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO platform_settings (key, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`, PlatformKeyFuelStationConfig, string(data))
	if err != nil {
		return FuelStationConfig{}, err
	}
	return cfg, nil
}

func GetFuelStationTemplateLibrary(ctx context.Context) ([]FuelStationTemplate, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var raw string
	err := db.QueryRowContext(ctx, `SELECT value FROM platform_settings WHERE key = $1`, PlatformKeyFuelStationTemplateLibrary).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultFuelStationTemplateLibrary(), nil
	}
	if err != nil {
		return nil, err
	}
	var out []FuelStationTemplate
	if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) == 0 {
		return defaultFuelStationTemplateLibrary(), nil
	}
	for i := range out {
		out[i].Payload = normalizeFuelStationConfig(out[i].Payload)
	}
	return out, nil
}

func UpdateFuelStationTemplateLibrary(ctx context.Context, templates []FuelStationTemplate) ([]FuelStationTemplate, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if len(templates) == 0 {
		templates = defaultFuelStationTemplateLibrary()
	}
	for i := range templates {
		templates[i].Payload = normalizeFuelStationConfig(templates[i].Payload)
	}
	data, err := json.Marshal(templates)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO platform_settings (key, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`, PlatformKeyFuelStationTemplateLibrary, string(data))
	if err != nil {
		return nil, err
	}
	return templates, nil
}

