# JobTracker — Phase 0 Launch Kit

Everything needed to run the Phase 0 pre-order validation: asset checklist,
ready-to-paste post copy per channel, launch sequence, and the go/no-go metric.

Landing page: `https://jobtracker.tecnologer.net` (Netlify, separate static repo).
Primary CTA: **Pre-order Pro — founder price** (Merchant-of-Record checkout).
Secondary CTA: email waitlist (Netlify Forms).

> Fill the placeholders before posting: `{{FOUNDER_PRICE}}` (e.g. $29),
> `{{LAUNCH_PRICE}}` (e.g. $39), `{{CHECKOUT_URL}}`, `{{REPO_URL}}`,
> `{{DEMO_URL}}`.

---

## 1. Pre-launch checklist (do before any posting)

- [ ] Landing page live on `jobtracker.tecnologer.net` with hero, privacy value
      prop, Community-vs-Pro table, founder price, pre-order + email CTAs, FAQ.
- [ ] Pre-order checkout live (Lemon Squeezy / Polar), refundable, clearly
      labelled "pre-order for an in-development app."
- [ ] Privacy analytics installed (Cloudflare Web Analytics or Plausible) and
      the funnel is visible: visits → CTA click → checkout start.
- [ ] Success metric written down (see §6) BEFORE launch.
- [ ] README updated with a "Pro coming soon → pre-order" line + landing link.
- [ ] Media assets ready (see §2).
- [ ] 3–5 people lined up to engage in the first hour of the HN / PH posts.

---

## 2. Media checklist (the asset kit)

One small kit covers every channel. Build these two things:

### A. Screenshots (needed everywhere) — must-have
Capture at a 1440×900 (or 1280×800) window, retina/2× if possible, PNG:

1. **Jobs table — light mode** (the main view, a few realistic rows).
2. **Jobs table — dark mode** (same view; dark mode is a selling point).
3. **Dashboard** — KPIs, status breakdown, stage funnel.
4. **Job detail dialog** — stages + stage history + a meeting (shows depth).
5. *(optional)* **Upcoming meetings** dropdown.

Specs: clean demo data (no test junk), consistent window size, hide personal
info. Keep 3 as the "hero" set (table light, dark, dashboard).

### B. One demo screen-recording (highest-ROI asset) — must-have
- 20–45s, no narration needed; just click through: add a job → move a stage →
  open detail → show dashboard → toggle dark mode.
- Export **two formats from the same recording**:
  - **GIF** (≤ ~8 MB, ~1000px wide) → Reddit, X, README, landing hero.
  - **MP4** (1080p) → Product Hunt, LinkedIn, X.
- Mac tools: **Kap** (free, records straight to GIF/MP4) or QuickTime + Gifski.

### C. Per-channel format cheat-sheet

| Channel | What to post |
|---|---|
| Hacker News (Show HN) | **Text only.** Media lives on the linked landing page. |
| r/selfhosted | Screenshots + demo GIF (image post outperforms text). |
| Product Hunt | Gallery images + 30–60s MP4 + logo/thumbnail. |
| dev.to / blog | Cover image + inline screenshots/GIFs. |
| LinkedIn | Native MP4 or image carousel; link in first comment. |
| X / #buildinpublic | Demo GIF or MP4 (autoplay motion wins). |
| awesome-selfhosted / alternativeto / Flathub | Clean screenshot + icon. |

> Don't over-produce. On HN/Reddit a raw, honest screen recording reads better
> than a glossy marketing video. A polished video only really pays off on PH.

---

## 3. Ready-to-paste copy per channel

### 3.1 Hacker News — Show HN
**Title (keep it plain; lead with local-first + open-source):**
```
Show HN: JobTracker – open-source, local-first job application tracker (Go/Vue)
```
**First comment:**
```
I got tired of job trackers that want your entire job search sitting on their
servers, so I built one that runs entirely on your machine.

JobTracker is a Go (stdlib mux + GORM + pure-Go SQLite) + Vue 3 app. It ships
two ways from one codebase: a self-hostable web app (Docker/K8s) and a native
desktop app (Wails) with no account and no network listener — your data is a
single local SQLite file you own.

It tracks applications, custom pipeline stages with full stage history,
contacts, meetings, and a dashboard (funnel + time-per-stage). CSV export, dark
mode. AGPL-3.0.

Demo: {{DEMO_URL}}  (login on the page)
Code: {{REPO_URL}}

I'm exploring a paid Pro tier (browser clipper, calendar/reminders, local
résumé↔JD match scoring) — feedback on whether that's useful, and on the
local-first approach generally, very welcome.
```
*Tone note: HN rewards the free/open-source tool and honesty; keep the Pro
mention low-key. Be online to answer comments — that's what sustains a Show HN.*

### 3.2 r/selfhosted
**Title:**
```
JobTracker – self-hostable, local-first job application tracker (Go/Vue, SQLite, Docker/K8s, no account)
```
**Body:**
```
Sharing a tool I built: a job-application tracker that's fully self-hostable and
local-first — no cloud, no account, your data is one SQLite file.

**Deploy:** Docker, docker-compose, or K8s (manifest included). Also a native
desktop build (Wails) for Linux/Windows/macOS, and a Flatpak.

**Features:** applications with custom pipeline stages + full stage history,
contacts, scheduled meetings with an upcoming view, a dashboard (status
breakdown, funnel, avg time per stage), CSV export, dark mode.

**Stack:** Go (stdlib mux, GORM, pure-Go SQLite) + Vue 3. AGPL-3.0.

[screenshots + short demo GIF here]

Demo: {{DEMO_URL}}
Code: {{REPO_URL}}

Full disclosure: the app is free and open-source; I'm testing interest in a paid
Pro tier for power features (browser clipper, calendar/reminders, local
résumé↔JD matching). Curious what this community thinks of that split.
```
*Read the sub's self-promo rules; engage with comments, don't drive-by.*

### 3.3 r/golang
**Title:**
```
I built an open-source, local-first job tracker in Go (stdlib mux + GORM + pure-Go SQLite, ships as web + Wails desktop)
```
**Body:** focus on the architecture — one codebase, two targets (web behind
Basic Auth vs Wails desktop with no listener), single-writer SQLite, GORM
AutoMigrate, `go:embed` of the built SPA. Link repo + invite code feedback.
*(Also usable, lightly adapted: r/vuejs for the frontend angle.)*

### 3.4 dev.to / blog post — outline
**Title:** "Why I built an open-source, local-first alternative to Teal and Huntr"
- The itch: every job tracker is a cloud SaaS that ingests your whole search.
- What local-first means here (SQLite file you own; desktop app, no listener).
- One Go codebase, two targets (web + Wails) — the interesting bits.
- Feature tour with screenshots/GIF.
- The open-core plan and the honest "will people pay?" question (link the
  landing page + pre-order).
- Ask for feedback.
*This article becomes shareable content you can repost everywhere, incl.
LinkedIn.*

### 3.5 LinkedIn (post as a story; link in FIRST comment)
```
Every job-application tracker I tried wanted my entire job search living on
their servers.

So I built the opposite: an open-source job tracker that runs entirely on your
machine. No account, no subscription, your data is a single file you own.

It handles applications, interview pipelines, contacts, meetings, and a
dashboard that shows where your search actually stands — and it ships as both a
self-hostable web app and a native desktop app.

I'm now testing whether people would pay for a Pro version. Building in public;
would love your take.

(demo + code in the comments)
```
First comment: `Demo: {{DEMO_URL}} · Code: {{REPO_URL}} · Pre-order Pro: {{CHECKOUT_URL}}`

### 3.6 X / Twitter (#buildinpublic) — thread
- **Tweet 1 (attach demo GIF):** "I built an open-source, local-first job
  tracker — no cloud, no account, your data stays on your machine. 🧵"
- **Tweet 2:** what it does (pipeline, meetings, dashboard, dark mode).
- **Tweet 3:** the stack (Go + Vue + SQLite; web + Wails desktop).
- **Tweet 4:** the open-core experiment + pre-order link + "would you pay?"
- Tags: #buildinpublic #golang #opensource #selfhosted

### 3.7 Newsletter pitch (selfh.st, Golang Weekly, Console.dev) — email
**Subject:** `Open-source, local-first job tracker (Go/Vue) — for a mention?`
```
Hi {{name}},

I built JobTracker, an open-source, local-first job-application tracker: Go +
Vue + SQLite, self-hostable via Docker/K8s with a native Wails desktop build.
No account, no cloud — your data is a single local file.

Thought it might fit {{newsletter}} given your {{self-hosted / Go}} audience.

Demo: {{DEMO_URL}}
Code: {{REPO_URL}}

Happy to answer anything. Thanks for considering it!
{{your name}}
```

### 3.8 Product Hunt
- **Name:** JobTracker
- **Tagline (≤60 chars):** `Own your job search — local-first, open-source tracker`
- **Description:** "JobTracker is an open-source, local-first job-application
  tracker. Track applications, interview pipelines, contacts, and meetings —
  with a dashboard that shows where your search stands. Self-host it or run the
  native desktop app. No account, no subscription: your data stays on your
  machine."
- **First comment:** the maker story (reuse the LinkedIn narrative) + pre-order
  link. Set up a "Coming Soon" page now to collect subscribers before launch.

---

## 4. Launch sequence

**Week 1 — soft launch (feedback + fixes):**
- [ ] Landing page + pre-order live; analytics confirmed.
- [ ] Share to your tecnologer audience (X, newsletter, YouTube/community).
- [ ] Post to 1–2 subreddits (r/golang or r/SideProject) framed as feedback.
- [ ] Publish the dev.to article.
- [ ] Fix whatever breaks; tighten the value prop from early reactions.

**Week 2 — coordinated spike (the real test):**
- [ ] Same day: Show HN + Product Hunt launch.
- [ ] Same day: r/selfhosted post (screenshots + GIF).
- [ ] Send newsletter pitches (selfh.st, Golang Weekly, Console.dev).
- [ ] LinkedIn story + X thread.
- [ ] Submit PRs to awesome-selfhosted / awesome-go; list on alternativeto;
      (optional) publish to Flathub.
- [ ] Be online all day to answer comments (this sustains HN/PH ranking).

*Concentrate traffic into one window so the funnel is readable.*

---

## 5. Where the ongoing (compounding) traffic comes from
Beyond the launch spike, these keep sending visitors for months:
awesome-selfhosted listing, Flathub presence, alternativeto entry, the dev.to
post's SEO, and the GitHub repo itself. Set these up once during Week 2.

---

## 6. Go/no-go metric (decide BEFORE launch)
Write your own numbers here and don't move them after the fact:

- Target relevant visitors in the 2–3 week window: **~1,000+**.
- **Go signal (build Phase 1):** e.g. **≥ 15–25 pre-orders**, or **≥ 3–5% of
  CTA-clickers reach checkout**.
- **Soft signal:** strong email-list growth + GitHub stars but few pre-orders →
  interest without willingness to pay → revisit price/positioning, don't build
  the full Pro yet.
- **No-go:** strong on-target traffic, negligible pre-orders → honest stop.
  Keep it a great free open-source tool + donations rather than building Pro.

The whole point of Phase 0: let real money (or its clear absence) make the
decision for you.
