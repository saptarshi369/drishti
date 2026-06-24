// Pure display formatters (kept out of components for readability).
export const usd = (n: number) => '$' + (n ?? 0).toFixed(2);
export const int = (n: number) => (n ?? 0).toLocaleString();
export const clock = (ms: number) => new Date(ms).toLocaleTimeString();

// ago renders a relative time like "2s ago" / "5m ago" / "3h ago" / "4d ago".
// ms is a Unix millisecond timestamp; the result is a human-readable string
// showing how long ago that moment was relative to now.
export const ago = (ms: number) => {
  const s = Math.max(0, Math.floor((Date.now() - ms) / 1000));
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
};

// tokensCompact renders a token count as a short human string: 950 -> "950",
// 34_000 -> "34k", 1_200_000 -> "1.2M".
export const tokensCompact = (n: number) => {
  const v = n ?? 0;
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(1).replace(/\.0$/, '') + 'M';
  if (v >= 1_000) return Math.round(v / 1_000) + 'k';
  return String(v);
};

// pct renders a 0-100 number as a whole-percent string, e.g. 41 -> "41%".
export const pct = (n: number) => `${Math.round(n ?? 0)}%`;
