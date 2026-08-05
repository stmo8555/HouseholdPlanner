const STATIC_CACHE = "hp-static-v51";

const OFFLINE_SHELL = "/static/offline-list.html";
const FALLBACK = "/static/offline-fallback.html";

const PRECACHE = [
    "/static/style.css",
    "/static/script.js",
    "/static/offline.js",
    "/static/offline-view.js",
    "/static/manifest.json",
    OFFLINE_SHELL,
    FALLBACK,
    "/static/vendor/htmx.min.js",
    "/static/vendor/fonts.css",
    "/static/vendor/roboto-normal-latin.woff2",
    "/static/vendor/roboto-normal-latin-ext.woff2",
    "/static/vendor/roboto-italic-latin.woff2",
    "/static/vendor/roboto-italic-latin-ext.woff2",
    "/static/assets/check.svg",
    "/static/assets/plus.svg",
    "/static/assets/send.svg",
    "/static/assets/shopping-cart.svg",
    "/static/assets/favicon-64.png",
    "/static/assets/apple-touch-icon.png",
    "/static/assets/background.jpg",
    "/static/assets/background-dark.jpg",
    "/static/assets/background-mobile.jpg",
    "/static/assets/icon-192.png",
    "/static/assets/icon-512.png",
    "/static/assets/icon-maskable-192.png",
    "/static/assets/icon-maskable-512.png",
];

self.addEventListener("install", event => {
    event.waitUntil(
        caches.open(STATIC_CACHE)
            .then(cache => cache.addAll(PRECACHE))
            .then(() => self.skipWaiting())
    );
});

self.addEventListener("activate", event => {
    event.waitUntil(
        caches.keys()
            .then(keys => Promise.all(
                keys.filter(k => k !== STATIC_CACHE).map(k => caches.delete(k))
            ))
            .then(() => self.clients.claim())
    );
});

async function staticCacheFirst(request) {
    const cached = await caches.match(request, { cacheName: STATIC_CACHE });
    if (cached) return cached;

    const resp = await fetch(request);
    if (resp.ok) {
        const cache = await caches.open(STATIC_CACHE);
        cache.put(request, resp.clone());
    }
    return resp;
}

self.addEventListener("fetch", event => {
    const request = event.request;
    if (request.method !== "GET") return;

    const url = new URL(request.url);
    if (url.origin !== self.location.origin) return;
    if (url.pathname === "/ping") return;

    // The offline shopping view must load with or without network.
    if (url.pathname === "/offline") {
        event.respondWith(
            caches.match(OFFLINE_SHELL, { cacheName: STATIC_CACHE })
                .then(cached => cached || fetch(request))
        );
        return;
    }

    if (url.pathname.startsWith("/static/") || url.pathname === "/sw.js") {
        event.respondWith(staticCacheFirst(request));
        return;
    }

    // Regular pages are online-only; a dead navigation gets the fallback.
    if (request.mode === "navigate") {
        event.respondWith(
            fetch(request).catch(() => caches.match(FALLBACK, { cacheName: STATIC_CACHE }))
        );
    }
});
