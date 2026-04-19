import { type KeyboardEvent } from 'react';
import { CirclePlus } from 'lucide-react';
import {
  activateTouchTap,
  beginTouchTapTracking,
  cancelTouchTapTracking,
  hasRecentTouchActivation,
  updateTouchTapTracking,
} from '@/lib/touch-activation';

export function CommandAddItem({
  query,
  onCreate,
}: {
  query: string;
  onCreate: () => void;
}) {
  return (
    <div
      role="option"
      aria-label={`Create "${query}"`}
      tabIndex={0}
      onClickCapture={(event) => {
        if (!hasRecentTouchActivation(event.currentTarget, event.timeStamp)) {
          return;
        }

        event.preventDefault();
        event.stopPropagation();
      }}
      onClick={onCreate}
      onTouchStart={(event) => {
        beginTouchTapTracking(event.currentTarget, event);
      }}
      onTouchMove={(event) => {
        updateTouchTapTracking(event.currentTarget, event);
      }}
      onTouchCancel={(event) => {
        cancelTouchTapTracking(event.currentTarget);
      }}
      onTouchEnd={(event) => {
        activateTouchTap(event.currentTarget, event);
      }}
      onKeyDown={(event: KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
          onCreate();
        }
      }}
      className="flex w-full cursor-pointer items-center rounded-sm px-2 py-1.5 focus:outline-none touch-manipulation"
    >
      <CirclePlus className="mr-2 h-4 w-4" />
      Create "{query}"
    </div>
  );
}
