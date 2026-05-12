import { useCallback, useEffect, useMemo, useState } from 'react';
import type { AnchorHTMLAttributes } from 'react';
import { Alert, Card, Spin, Tabs, Typography } from 'antd';
import type { Components } from 'react-markdown';
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeSanitize from 'rehype-sanitize';
import remarkGfm from 'remark-gfm';
import 'highlight.js/styles/github.css';

import { ideMarkdownSanitizeSchema } from './ideMarkdownSanitizeSchema';
import styles from './DeveloperIdeClientsPage.module.css';

const { Title, Paragraph, Text } = Typography;

const DOC_BASE = '/docs/developer';

const DOC_TABS = [
  { key: 'overview', label: '总览', file: 'ide-overview.md' },
  { key: 'claude-code', label: 'Claude Code', file: 'ide-claude-code.md' },
  { key: 'cursor', label: 'Cursor', file: 'ide-cursor.md' },
  { key: 'opencode', label: 'OpenCode', file: 'ide-opencode.md' },
  { key: 'codex', label: 'Codex', file: 'ide-codex.md' },
  { key: 'cline', label: 'Cline', file: 'ide-cline.md' },
  { key: 'windsurf', label: 'Windsurf', file: 'ide-windsurf.md' },
  { key: 'trae', label: 'Trae', file: 'ide-trae.md' },
  { key: 'codebuddy', label: 'CodeBuddy', file: 'ide-codebuddy.md' },
] as const;

async function fetchMarkdown(file: string): Promise<string> {
  const res = await fetch(`${DOC_BASE}/${file}`);
  if (!res.ok) throw new Error(`${file}: ${res.status}`);
  return res.text();
}

/** highlight 在前、sanitize 在后，以保留 hljs 结构并剔除脚本/危险属性 */
const rehypePlugins: import('react-markdown').Options['rehypePlugins'] = [
  [rehypeHighlight, { detect: true }],
  [rehypeSanitize, ideMarkdownSanitizeSchema],
];

/** 开发者中心：IDE / CLI 分 Tab Markdown（源文件在 public/docs/developer/ide-*.md） */
export default function DeveloperIdeClientsPage() {
  const [activeKey, setActiveKey] = useState<string>(DOC_TABS[0].key);
  const [cache, setCache] = useState<Record<string, string>>({});
  const [loadingKey, setLoadingKey] = useState<string | null>(DOC_TABS[0].key);
  const [error, setError] = useState<string | null>(null);

  const loadTab = useCallback(async (key: string) => {
    const tab = DOC_TABS.find((t) => t.key === key);
    if (!tab) return;
    setError(null);
    setLoadingKey(key);
    try {
      const md = await fetchMarkdown(tab.file);
      setCache((prev) => ({ ...prev, [key]: md }));
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : `无法加载 ${tab.file}，请确认 public/docs/developer 下存在该文件且已构建。`
      );
    } finally {
      setLoadingKey(null);
    }
  }, []);

  useEffect(() => {
    if (cache[activeKey]) return;
    void loadTab(activeKey);
  }, [activeKey, cache, loadTab]);

  const markdownComponents = useMemo<Components>(
    () => ({
      a: ({ href, children, ...rest }: AnchorHTMLAttributes<HTMLAnchorElement>) => {
        const ext = href?.startsWith('http');
        return (
          <a
            href={href}
            {...rest}
            target={ext ? '_blank' : undefined}
            rel={ext ? 'noreferrer' : undefined}
          >
            {children}
          </a>
        );
      },
    }),
    []
  );

  return (
    <div>
      <Title level={3} style={{ marginTop: 0 }}>
        IDE 与 CLI 接入
      </Title>
      <Paragraph type="secondary">
        按工具分册；正文为 <Text code>react-markdown</Text> 渲染（标题、表格、<Text code>```</Text>{' '}
        代码高亮），并经 <Text code>rehype-sanitize</Text> 按 GitHub 风格白名单清洗以降低 XSS
        风险。源文件位于 <Text code>public/docs/developer/ide-*.md</Text>。
      </Paragraph>

      {error && (
        <Alert
          type="error"
          message={error}
          style={{ marginBottom: 16 }}
          showIcon
          closable
          onClose={() => setError(null)}
        />
      )}

      <Tabs
        activeKey={activeKey}
        onChange={(k) => setActiveKey(k)}
        destroyInactiveTabPane
        type="card"
        size="small"
        tabBarStyle={{ marginBottom: 12 }}
        items={DOC_TABS.map((t) => {
          const md = cache[t.key];
          const tabLoading = loadingKey === t.key && !md;
          return {
            key: t.key,
            label: t.label,
            children: (
              <Card size="small" styles={{ body: { padding: '16px 20px' } }}>
                <div className={styles.panel}>
                  {tabLoading ? (
                    <Spin />
                  ) : md ? (
                    <div className={styles.markdownBody}>
                      <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        rehypePlugins={rehypePlugins}
                        components={markdownComponents}
                      >
                        {md}
                      </ReactMarkdown>
                    </div>
                  ) : null}
                </div>
              </Card>
            ),
          };
        })}
      />
    </div>
  );
}
