# Job Tracker — Icon Assets

Final app icon ("Loading… hired"): a blueprint-style briefcase over a progress bar, on a dark indigo tile. Colors: background `#0F1126`, line work `#8B93FF`, progress fill `#64E3F0`, track `#252A58`.

## Files

| File | Size | Usage |
|---|---|---|
| `icon.svg` | vector | Master source. Use anywhere that accepts SVG (modern favicons via `<link rel="icon" type="image/svg+xml">`, docs, README). Edit this file to regenerate the PNGs. |
| `icon-512.png` | 512×512 | PWA manifest icon (`"sizes": "512x512"`), app store listings, any large raster need. |
| `icon-192.png` | 192×192 | PWA manifest icon (`"sizes": "192x192"`) — Android home screen. |
| `apple-touch-icon.png` | 180×180 | iOS home-screen icon. Reference as `<link rel="apple-touch-icon" href="/apple-touch-icon.png">`. |
| `favicon-32.png` | 32×32 | Standard browser-tab favicon. `<link rel="icon" sizes="32x32" href="/favicon-32.png">`. |
| `favicon-16.png` | 16×16 | Small favicon for dense tab bars / bookmarks. `<link rel="icon" sizes="16x16" href="/favicon-16.png">`. |
| `social-banner.png` | 1280×640 | GitHub social preview (repo Settings → Social preview) and Open Graph image (`og:image`, `twitter:image`). |

## Setup (Vue SPA)

Copy the PNGs to `web/public/`, then in `web/index.html`:

```html
<link rel="icon" href="/favicon-32.png" sizes="32x32">
<link rel="icon" href="/favicon-16.png" sizes="16x16">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
```

If you add a PWA manifest (`manifest.webmanifest`):

```json
{
  "icons": [
    { "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/icon-512.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```
