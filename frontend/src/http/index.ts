import axios from "axios";

const CACHE_KEY = "http_cache";
const CACHE_LIMIT = 50;
const CACHE_TTL = 60 * 1000; // 1 minute

interface CacheEntry {
  url: string;
  method: string;
  timestamp: number;
  data: any;
}

const CACHE_BYPASS_PATTERNS: RegExp[] = [
  /^\/api\/v1\/sheets\/get(\?|$)/,
];

const CACHE_BYPASS_HEADER = "x-cache-bypass";

const normalizeUrl = (url: string) => {
  try {
    const resolved = new URL(url, window.location.origin);
    resolved.searchParams.delete("_nocache");
    return `${resolved.pathname}${resolved.search}`;
  } catch {
    return url;
  }
};

const shouldBypassCache = (url: string) => {
  if (!url) return false;
  const normalized = normalizeUrl(url);
  return CACHE_BYPASS_PATTERNS.some((pattern) => pattern.test(normalized));
};

const safeParseCache = (raw: string | null): CacheEntry[] => {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((entry) => (
      entry &&
      typeof entry.url === "string" &&
      typeof entry.method === "string" &&
      typeof entry.timestamp === "number"
    ));
  } catch {
    return [];
  }
};

const getCache = (): CacheEntry[] => {
  return safeParseCache(localStorage.getItem(CACHE_KEY));
};

const setCache = (cache: CacheEntry[]) => {
  localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
};

const addToCache = (url: string, method: string, data: any) => {
  if (shouldBypassCache(url)) return;
  const normalizedUrl = normalizeUrl(url);
  const cache = getCache();
  const existingIndex = cache.findIndex(entry => entry.url === normalizedUrl && entry.method === method);

  if (existingIndex !== -1) {
    cache[existingIndex] = { url: normalizedUrl, method, timestamp: Date.now(), data };
  } else {
    cache.push({ url: normalizedUrl, method, timestamp: Date.now(), data });
    if (cache.length > CACHE_LIMIT) {
      cache.shift(); // Remove oldest
    }
  }

  setCache(cache);
};

const getFromCache = (url: string, method: string): any | null => {
  if (shouldBypassCache(url)) return null;
  const normalizedUrl = normalizeUrl(url);
  const cache = getCache();
  const entry = cache.find(entry => entry.url === normalizedUrl && entry.method === method);

  if (!entry) return null;
  if (Date.now() - entry.timestamp >= CACHE_TTL) return null;
  return entry.data;
};

const http = axios.create({
  baseURL: "/",
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.request.use(async (config) => {
  const session = localStorage.getItem("session");
  if (session) {
    config.headers["Authorization"] = `Bearer ${session}`;
  }

  if (config.method?.toLowerCase() === "get") {
    const url = config.url || "";
    const bypassHeader = Boolean(config.headers?.[CACHE_BYPASS_HEADER]);
    if (!bypassHeader && !shouldBypassCache(url)) {
      const cachedData = getFromCache(url, "get");
      if (cachedData) {
        const error: any = new Error("CACHED_RESPONSE");
        error.config = config;
        error.cachedData = cachedData;
        throw error;
      }
    }
  }

  return config;
});

http.interceptors.response.use(
  (response) => {
    if (response.config.method?.toLowerCase() === "get") {
      const url = response.config.url || "";
      const bypassHeader = Boolean(response.config.headers?.[CACHE_BYPASS_HEADER]);
      if (!bypassHeader && !shouldBypassCache(url)) {
        addToCache(url, "get", response.data);
      }
    }
    return response;
  }, 
  async (error) => {
    if (error.message === "CACHED_RESPONSE" && error.cachedData) {
      return {
        data: error.cachedData,
        status: 200,
        statusText: "OK",
        headers: {
          "x-client-cache": "hit",
        },
        config: error.config,
        cached: true
      };
    }
    
    if (error.config && error.config.method?.toLowerCase() === "get") {
      const url = error.config.url || "";
      const bypassHeader = Boolean(error.config.headers?.[CACHE_BYPASS_HEADER]);
      if (!bypassHeader && !shouldBypassCache(url)) {
        const cachedData = getFromCache(url, "get");
        if (cachedData) {
          return {
            data: cachedData,
            status: 200,
            statusText: "OK",
            headers: {
              "x-client-cache": "stale",
            },
            config: error.config,
            cached: true
          };
        }
      }
    }
    
    return Promise.reject(error);
  }
);

export default http;