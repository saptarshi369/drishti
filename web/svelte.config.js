import adapter from '@sveltejs/adapter-static';

export default {
  kit: {
    // SPA mode: one index.html fallback, served by Go for any route.
    adapter: adapter({ fallback: 'index.html' }),
    prerender: { entries: [] }
  }
};
