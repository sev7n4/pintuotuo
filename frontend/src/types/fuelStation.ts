export interface FuelStationTierConfig {
  label: string;
  sku_id: number;
}

export interface FuelStationSectionConfig {
  code: string;
  name: string;
  description: string;
  badge: string;
  sort_order: number;
  status: 'active' | 'inactive';
  tiers: FuelStationTierConfig[];
}

export interface FuelStationConfig {
  page_title: string;
  page_subtitle: string;
  rule_text: string;
  sections: FuelStationSectionConfig[];
}

export interface FuelStationTemplate {
  key: string;
  name: string;
  description: string;
  payload: FuelStationConfig;
}
