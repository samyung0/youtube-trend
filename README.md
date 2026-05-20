# ThumbTrend - YouTube Thumbnail Trend Analyzer

AI-powered tool that tracks trending YouTube thumbnails, discovers micro-genres, and analyzes visual patterns (colors, faces, text) across categories.

## Features

- **Trending Dashboard** — Live grid of trending YouTube thumbnails with stats
- **AI Micro-Genres** — LLM-powered genre clustering (e.g. "Cozy Farming Sims", "AI Coding Tutorials")
- **Thumbnail Analysis** — Dominant colors, face detection, OCR text extraction, brightness
- **Historical Trends** — Track genre popularity over days/weeks with charts
- **Time Filtering** — 24h / 7d / 30d views

## Tech Stack

- **Frontend**: Next.js 14 (App Router), TailwindCSS, shadcn/ui, Recharts
- **Database**: Supabase (PostgreSQL)
- **Scraper**: Node.js scripts running on GitHub Actions cron
- **AI**: DeepSeek V4 Flash for genre clustering (OpenAI-compatible API)
- **Image Analysis**: Sharp (colors), Tesseract.js (OCR), skin-tone heuristic (faces)

## Setup

### 1. Create External Services

| Service | URL | What You Need |
|---------|-----|---------------|
| Supabase | https://supabase.com | `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_KEY` |
| YouTube API | https://console.cloud.google.com | Enable "YouTube Data API v3", create API key |
| DeepSeek | https://platform.deepseek.com | API key for deepseek-v4-flash |

### 2. Database Setup

Run the migration SQL in your Supabase SQL Editor:

```
supabase/migrations/001_initial_schema.sql
```

### 3. Environment Variables

```bash
cp .env.local.example .env.local
# Fill in your keys
```

### 4. Install & Run

```bash
npm install
npm run dev
```

### 5. Scraper (Local Test)

```bash
cd scripts/scraper
npm install
# Set env vars then:
npx tsx fetch-trending.ts    # Fetch trending videos
npx tsx cluster-genres.ts    # AI genre clustering
npx tsx analyze-thumbnails.ts # Thumbnail analysis
```

### 6. GitHub Actions (Automated)

Add these secrets to your repo (Settings > Secrets > Actions):

- `YOUTUBE_API_KEY`
- `SUPABASE_URL`
- `SUPABASE_SERVICE_KEY`
- `DEEPSEEK_API_KEY`

The workflow runs every 4 hours automatically. You can also trigger it manually from the Actions tab.

## API Quota

YouTube API: ~60 units/day out of 10,000 free limit (0.6%).
DeepSeek: ~$0.003/day for genre clustering ($0.14/1M input tokens).

## Project Structure

```
├── .github/workflows/scrape.yml  # Cron scraper
├── scripts/scraper/
│   ├── fetch-trending.ts          # YouTube API fetcher
│   ├── cluster-genres.ts          # LLM genre clustering
│   ├── analyze-thumbnails.ts      # Image analysis
│   └── lib/                       # Shared scraper utilities
├── src/
│   ├── app/                       # Next.js pages
│   │   ├── page.tsx               # Dashboard
│   │   ├── genre/[slug]/page.tsx  # Genre detail
│   │   ├── trends/page.tsx        # Historical trends
│   │   └── analysis/page.tsx      # Thumbnail analysis
│   ├── components/                # React components
│   └── lib/                       # Supabase client, queries, types
└── supabase/migrations/           # Database schema
```
# youtube-trend
