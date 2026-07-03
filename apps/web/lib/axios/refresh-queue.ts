type QueueEntry = {
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
};

let isRefreshing = false;
let failedQueue: QueueEntry[] = [];

export function getIsRefreshing() {
  return isRefreshing;
}

export function setIsRefreshing(value: boolean) {
  isRefreshing = value;
}

export function pushToQueue(entry: QueueEntry) {
  failedQueue.push(entry);
}

export const processQueue = (error: unknown) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve();
    }
  });
  failedQueue = [];
};
