import { PropsWithChildren } from 'react';
import { AlertCircle } from 'lucide-react';
import clsx from 'clsx';

import { Icon, IconMode } from '@@/Icon';

type Color = 'orange' | 'blue';

export interface Props {
  icon?: React.ReactNode;
  color?: Color;
  className?: string;
  inline?: boolean;
}

export function TextTip({
  color = 'orange',
  icon = AlertCircle,
  inline = true,
  className,
  children,
}: PropsWithChildren<Props>) {
  return (
    <div
      className={clsx(
        className,
        'small items-center gap-1',
        inline ? 'inline-flex' : 'flex'
      )}
    >
      <Icon icon={icon} mode={getMode(color)} />

      <span className="text-muted">{children}</span>
    </div>
  );
}

function getMode(color: Color): IconMode {
  switch (color) {
    case 'blue':
      return 'primary';
    case 'orange':
    default:
      return 'warning';
  }
}
