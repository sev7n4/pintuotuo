import type { Schema } from 'hast-util-sanitize';
import { defaultSchema } from 'rehype-sanitize';

/**
 * GitHub 风格 defaultSchema + rehype-highlight 所需的 class（hljs / language-* / token span）。
 */
export const ideMarkdownSanitizeSchema: Schema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    code: [...(defaultSchema.attributes?.code ?? []), ['className', 'hljs']],
    span: [...(defaultSchema.attributes?.span ?? []), ['className', /^hljs-/]],
  },
};
