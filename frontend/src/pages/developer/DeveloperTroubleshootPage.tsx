import { useEffect, useState } from 'react';
import { Alert, Card, Spin, Tabs, Typography } from 'antd';

const { Title, Paragraph } = Typography;

async function fetchText(path: string): Promise<string> {
  const res = await fetch(path);
  if (!res.ok) throw new Error(String(res.status));
  return res.text();
}

export default function DeveloperTroubleshootPage() {
  const [errorsMd, setErrorsMd] = useState<string>('');
  const [routingMd, setRoutingMd] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setErr(null);
      try {
        const [e, r] = await Promise.all([
          fetchText('/docs/developer/errors.md'),
          fetchText('/docs/developer/routing-ssot.md'),
        ]);
        if (!cancelled) {
          setErrorsMd(e);
          setRoutingMd(r);
        }
      } catch {
        if (!cancelled) setErr('无法加载文档，请确认已构建且 public/docs 存在。');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div>
      <Title level={3} style={{ marginTop: 0 }}>
        错误与排障
      </Title>
      <Paragraph type="secondary">
        下列内容与前端 axios 拦截器、后端代理状态映射保持一致；更新实现时请同步修改{' '}
        <Typography.Text code>public/docs/developer/*.md</Typography.Text>。IDE / CLI 总览见{' '}
        <a href="/developer/ide-clients">IDE 与 CLI 接入</a>。
      </Paragraph>

      {err && <Alert type="error" message={err} style={{ marginBottom: 16 }} />}

      {loading ? (
        <Spin />
      ) : (
        <Card size="small">
          <Tabs
            items={[
              {
                key: 'errors',
                label: 'HTTP 与鉴权',
                children: (
                  <pre
                    style={{
                      margin: 0,
                      whiteSpace: 'pre-wrap',
                      fontSize: 13,
                      maxHeight: '70vh',
                      overflow: 'auto',
                    }}
                  >
                    {errorsMd}
                  </pre>
                ),
              },
              {
                key: 'routing',
                label: '路由与出站（摘要）',
                children: (
                  <pre
                    style={{
                      margin: 0,
                      whiteSpace: 'pre-wrap',
                      fontSize: 13,
                      maxHeight: '70vh',
                      overflow: 'auto',
                    }}
                  >
                    {routingMd}
                  </pre>
                ),
              },
            ]}
          />
        </Card>
      )}
    </div>
  );
}
