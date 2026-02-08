type WaitForPdfOptions = {
  retries?: number;
  intervalMs?: number;
  onAttempt?: (attempt: number, maxAttempts: number) => void;
};

export const resolvePdfUrl = (url: string) => {
  if (!url) return "";
  if (url.startsWith("http://") || url.startsWith("https://")) return url;
  if (url.startsWith("/")) return url;
  return `/${url}`;
};

export const checkPdfAvailable = async (url: string) => {
  const resolved = resolvePdfUrl(url);
  if (!resolved) return false;

  try {
    const headResponse = await fetch(resolved, { method: "HEAD", cache: "no-store" });
    if (headResponse.ok) {
      return true;
    }

    if (headResponse.status !== 405 && headResponse.status !== 403 && headResponse.status !== 404) {
      return false;
    }

    const rangeResponse = await fetch(resolved, {
      method: "GET",
      cache: "no-store",
      headers: {
        Range: "bytes=0-0",
      },
    });

    return rangeResponse.ok || rangeResponse.status === 206;
  } catch {
    return false;
  }
};

export const waitForPdfReady = async (url: string, options?: WaitForPdfOptions) => {
  const retries = options?.retries ?? 8;
  const intervalMs = options?.intervalMs ?? 1500;
  const maxAttempts = retries + 1;

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    if (attempt > 1) {
      await new Promise((resolve) => setTimeout(resolve, intervalMs));
    }

    const available = await checkPdfAvailable(url);
    if (available) {
      return true;
    }

    options?.onAttempt?.(attempt, maxAttempts);
  }

  return false;
};
