const RECENT_TOUCH_ACTIVATION_ATTR = 'data-recent-touch-activation';
const TOUCH_START_X_ATTR = 'data-touch-start-x';
const TOUCH_START_Y_ATTR = 'data-touch-start-y';
const TOUCH_MOVED_ATTR = 'data-touch-moved';
const TOUCH_TAP_MAX_DISTANCE_PX = 10;
const RECENT_TOUCH_ACTIVATION_MS = 750;

type TouchPoint = {
  clientX: number;
  clientY: number;
};

type TouchListLike = {
  0?: TouchPoint;
  length: number;
};

type TouchLikeEvent = {
  touches: TouchListLike;
  changedTouches: TouchListLike;
};
type TouchTapEndEvent = TouchLikeEvent &
  Pick<TouchEvent, 'preventDefault' | 'timeStamp'>;

function isElement(target: EventTarget | null): target is HTMLElement {
  return target instanceof HTMLElement;
}

function readTouchPoint(event: TouchLikeEvent) {
  const touch = event.changedTouches[0] ?? event.touches[0];

  if (!touch) {
    return null;
  }

  return { x: touch.clientX, y: touch.clientY };
}

function clearTouchTapTracking(target: HTMLElement) {
  target.removeAttribute(TOUCH_START_X_ATTR);
  target.removeAttribute(TOUCH_START_Y_ATTR);
  target.removeAttribute(TOUCH_MOVED_ATTR);
}

function readTouchStart(target: HTMLElement) {
  const startX = Number(target.getAttribute(TOUCH_START_X_ATTR));
  const startY = Number(target.getAttribute(TOUCH_START_Y_ATTR));

  if (!Number.isFinite(startX) || !Number.isFinite(startY)) {
    return null;
  }

  return { x: startX, y: startY };
}

function exceededTapDistance(
  start: { x: number; y: number },
  current: { x: number; y: number }
) {
  return Math.hypot(current.x - start.x, current.y - start.y) > TOUCH_TAP_MAX_DISTANCE_PX;
}

export function hasRecentTouchActivation(
  target: EventTarget | null,
  timeStamp: number
) {
  if (!isElement(target)) {
    return false;
  }

  const lastTouchActivation = Number(
    target.getAttribute(RECENT_TOUCH_ACTIVATION_ATTR)
  );

  return (
    Number.isFinite(lastTouchActivation) &&
    timeStamp - lastTouchActivation < RECENT_TOUCH_ACTIVATION_MS
  );
}

export function markRecentTouchActivation(
  target: EventTarget | null,
  timeStamp: number
) {
  if (!isElement(target)) {
    return;
  }

  target.setAttribute(RECENT_TOUCH_ACTIVATION_ATTR, String(timeStamp));
}

export function beginTouchTapTracking(
  target: EventTarget | null,
  event: TouchLikeEvent
) {
  if (!isElement(target)) {
    return;
  }

  const point = readTouchPoint(event);

  if (!point) {
    return;
  }

  target.setAttribute(TOUCH_START_X_ATTR, String(point.x));
  target.setAttribute(TOUCH_START_Y_ATTR, String(point.y));
  target.setAttribute(TOUCH_MOVED_ATTR, 'false');
}

export function updateTouchTapTracking(
  target: EventTarget | null,
  event: TouchLikeEvent
) {
  if (!isElement(target)) {
    return;
  }

  const point = readTouchPoint(event);
  const start = readTouchStart(target);

  if (!point || !start) {
    return;
  }

  if (exceededTapDistance(start, point)) {
    target.setAttribute(TOUCH_MOVED_ATTR, 'true');
  }
}

export function cancelTouchTapTracking(target: EventTarget | null) {
  if (!isElement(target)) {
    return;
  }

  clearTouchTapTracking(target);
}

export function finishTouchTapTracking(
  target: EventTarget | null,
  event: TouchLikeEvent
) {
  if (!isElement(target)) {
    return false;
  }

  const point = readTouchPoint(event);
  const start = readTouchStart(target);
  const moved = target.getAttribute(TOUCH_MOVED_ATTR) === 'true';

  clearTouchTapTracking(target);

  if (!point || !start || moved) {
    return false;
  }

  return !exceededTapDistance(start, point);
}

export function activateTouchTap(
  target: EventTarget | null,
  event: TouchTapEndEvent
) {
  if (!finishTouchTapTracking(target, event)) {
    return false;
  }

  if (!isElement(target)) {
    return false;
  }

  event.preventDefault();
  target.click();
  markRecentTouchActivation(target, event.timeStamp);
  return true;
}
