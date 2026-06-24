// Theme + accent toggles. Sets data-theme/data-accent on <html> and persists to
// localStorage so a reload keeps the user's choice. One module owns this state.
import { writable } from 'svelte/store';

type Theme = 'dark' | 'light';
type Accent = 'indigo' | 'teal' | 'violet';

// readLS: safely reads from localStorage; returns fallback when unavailable (SSR).
const readLS = (k: string, fallback: string) =>
  (typeof localStorage !== 'undefined' && localStorage.getItem(k)) || fallback;

// timeOfDayTheme: the default theme when the user has NOT picked one explicitly.
// Light during daytime (07:00–18:59 local), dark in the evening/night. Keep the
// 7/19 boundary in sync with the inline pre-paint script in app.html.
export function timeOfDayTheme(): Theme {
  const h = new Date().getHours();
  return h >= 7 && h < 19 ? 'light' : 'dark';
}

// storedTheme: the user's explicit choice, or null if they have never toggled.
// We must distinguish "no choice yet" (→ follow the clock) from a real choice,
// so we read the raw key rather than using readLS's non-null fallback.
const storedTheme = (): Theme | null => {
  if (typeof localStorage === 'undefined') return null;
  const v = localStorage.getItem('theme');
  return v === 'dark' || v === 'light' ? v : null;
};

// Initial theme: the user's stored choice if any, otherwise time-of-day.
export const theme = writable<Theme>(storedTheme() ?? timeOfDayTheme());
export const accent = writable<Accent>(readLS('accent', 'indigo') as Accent);

/** applyTheme sets data-theme on <html>. When persist is true (an explicit user
 *  selection via the toggle/settings) it also records the choice in localStorage.
 *  The initial auto-apply passes persist=false so a time-of-day default is NOT
 *  frozen and re-evaluates against the clock on the next visit. */
export function applyTheme(t: Theme, persist = true) {
  document.documentElement.setAttribute('data-theme', t);
  if (persist) localStorage.setItem('theme', t);
}

/** applyAccent: sets or removes data-accent on <html>; indigo is the default
 *  token set so no attribute is needed for it. Persists to localStorage. */
export function applyAccent(a: Accent) {
  // indigo is the default token set → no data-accent attribute needed.
  if (a === 'indigo') document.documentElement.removeAttribute('data-accent');
  else document.documentElement.setAttribute('data-accent', a);
  localStorage.setItem('accent', a);
}
