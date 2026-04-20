import { type HTMLAttributes, type KeyboardEvent } from 'react';
import { CirclePlus } from 'lucide-react';
import {
  activateTouchTap,
  beginTouchTapTracking,
  cancelTouchTapTracking,
  shouldSuppressTouchClick,
  updateTouchTapTracking,
} from '@/lib/touch-activation';

type TouchActivationHandlers = Pick<
  HTMLAttributes<HTMLDivElement>,
  'onClickCapture' | 'onTouchCancel' | 'onTouchEnd' | 'onTouchMove' | 'onTouchStart'
>;

export function CommandAddItem({
  query,
  onCreate,
  touchEnabled = false,
}: {
  query: string;
  onCreate: () => void;
  touchEnabled?: boolean;
}) {
  const touchActivationHandlers: TouchActivationHandlers = touchEnabled
    ? {
        onClickCapture: (event) => {
          if (!shouldSuppressTouchClick(event.currentTarget, event.timeStamp)) {
            return;
          }

          event.preventDefault();
          event.stopPropagation();
        },
        onTouchStart: (event) => {
          beginTouchTapTracking(event.currentTarget, event);
        },
        onTouchMove: (event) => {
          updateTouchTapTracking(event.currentTarget, event);
        },
        onTouchCancel: (event) => {
          cancelTouchTapTracking(event.currentTarget);
        },
        onTouchEnd: (event) => {
          activateTouchTap(event.currentTarget, event);
        },
      }
    : {};

  return (
    <div
      role="option"
      aria-label={`Create "${query}"`}
      tabIndex={0}
      onClick={onCreate}
      {...touchActivationHandlers}
      onKeyDown={(event: KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
          onCreate();
        }
      }}
      className={`flex w-full cursor-pointer items-center rounded-sm px-2 py-1.5 focus:outline-none${touchEnabled ? ' touch-manipulation' : ''}`}
    >
      <CirclePlus className="mr-2 h-4 w-4" />
      Create "{query}"
    </div>
  );
}
