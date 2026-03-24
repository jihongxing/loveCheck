/**
 * Lightweight analytics helper.
 * Sends custom events to Umami (if loaded) via window.umami.track().
 * No-ops silently when analytics is not configured.
 */
export function trackEvent(eventName, data) {
  try {
    if (window.umami && typeof window.umami.track === 'function') {
      window.umami.track(eventName, data)
    }
  } catch {
    // analytics should never break the app
  }
}
