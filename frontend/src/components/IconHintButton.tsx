import { Button, Tooltip } from 'antd';
import type { ButtonProps } from 'antd';

export type IconHintButtonProps = Omit<ButtonProps, 'title'> & {
  /** Tooltip 文案；默认用作 aria-label（可被 aria-label 覆盖） */
  hint: string;
};

/**
 * 以图标为主、悬停展示完整说明；禁用时仍可通过 Tooltip 查看原因。
 */
export function IconHintButton({
  hint,
  disabled,
  block,
  className,
  style,
  'aria-label': ariaLabel,
  ...rest
}: IconHintButtonProps) {
  const label = ariaLabel ?? hint;
  const btn = (
    <Button
      {...rest}
      disabled={disabled}
      block={block}
      className={className}
      style={style}
      aria-label={label}
    />
  );
  const inner = disabled ? (
    <span
      style={
        block
          ? { display: 'flex', width: '100%', minWidth: 0, cursor: 'not-allowed' }
          : { display: 'inline-flex', cursor: 'not-allowed' }
      }
    >
      {btn}
    </span>
  ) : (
    btn
  );
  return <Tooltip title={hint}>{inner}</Tooltip>;
}
