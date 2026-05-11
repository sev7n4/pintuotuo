export type CatalogBannerShowWhen = 'always' | 'group_only';

export interface CatalogGroupBannerConfig {
  enabled: boolean;
  show_when: CatalogBannerShowWhen;
  tag?: string;
  title: string;
  subtitle?: string;
  cta_label?: string;
  /** 站内相对路径或绝对 URL */
  cta_href?: string;
  image?: string;
  background?: string;
}

export function defaultCatalogGroupBanner(): CatalogGroupBannerConfig {
  return {
    enabled: false,
    show_when: 'group_only',
    title: '',
  };
}
