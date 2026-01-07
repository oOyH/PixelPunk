// Runtime config for PixelPunk web (loaded before the app bundle).
// In production (Docker all-in-one), this is served dynamically by the backend at `/runtime-config.js`.
// In local `vite` dev, this static file prevents a 404 and provides sensible defaults.

;(function () {
  const existing = window.__VITE_RUNTIME_CONFIG__ && typeof window.__VITE_RUNTIME_CONFIG__ === 'object' ? window.__VITE_RUNTIME_CONFIG__ : {}

  window.__VITE_RUNTIME_CONFIG__ = {
    VITE_API_BASE_URL: '/api/v1',
    VITE_SITE_DOMAIN: '',
    ...existing,
  }

  if (!window.__VITE_API_BASE_URL__) {
    window.__VITE_API_BASE_URL__ = window.__VITE_RUNTIME_CONFIG__.VITE_API_BASE_URL
  }

  if (!window.__VITE_SITE_DOMAIN__) {
    window.__VITE_SITE_DOMAIN__ = window.__VITE_RUNTIME_CONFIG__.VITE_SITE_DOMAIN
  }
})()

