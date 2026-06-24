// Theme + accent toggles. Sets data-theme/data-accent on <html> and persists to
// localStorage so a reload keeps the user's choice. One module owns this state.
import { writable } from 'svelte/store';

type Theme = 'dark' | 'light';
type Accent = 'indigo' | 'teal' | 'violet';

// readLS: safely reads from localStorage; returns fallback when unavailable (SSR).
const readLS = (k: string, fallback: string) =>
  (typeof localStorage !== 'undefined' && localStorage.getItem(k)) || fallback;

export const theme = writable<Theme>(readLS('theme', 'dark') as Theme);
export const accent = writable<Accent>(readLS('accent', 'indigo') as Accent);

/** applyTheme: sets data-theme on <html> and persists the choice. */
export function applyTheme(t: Theme) {
  document.documentElement.setAttribute('data-theme', t);
  localStorage.setItem('theme', t);
}

/** applyAccent: sets or removes data-accent on <html>; indigo is the default
 *  token set so no attribute is needed for it. Persists to localStorage. */
export function applyAccent(a: Accent) {
  // indigo is the default token set → no data-accent attribute needed.
  if (a === 'indigo') document.documentElement.removeAttribute('data-accent');
  else document.documentElement.setAttribute('data-accent', a);
  localStorage.setItem('accent', a);
}
