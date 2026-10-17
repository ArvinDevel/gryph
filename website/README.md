# Gryph Website

## Development

> Install and setup [PMG](https://github.com/safedep/pmg) before you install dependencies.

Install dependencies:

```
npm install
```

Start development server:

```
npm run dev
```

## Deployment

Build static site:

```
npm run build
```

Deploy to Cloudflare Pages:

```
npx wrangler deploy
```

## Design System

The site uses the Cartridge design system. Find the styles in `src/shared/`:

- `tokens.css` holds the colours, fonts, borders and shadows. Do not write a colour value outside this file. Build each shadow from these tokens.
- Page styles use the role tokens (`--bg`, `--surface`, `--fg`, `--rule`, `--shadow-c` and the others). The dark theme changes only the role tokens and the soft and deep accents.
- The raw ink scale (`--ink-900` and the others) stays the same in both themes. Use it only on surfaces that are always dark, such as terminal windows and the ink band.
- Text on a coin, mint, cherry, sky or grape fill uses `--on-accent`.
- `base.css` holds the reset and the global element styles.
- `components.css` holds the shared parts: buttons, cards, windows, tabs, commands, agents and the footer.

Each page imports `components.css` and adds only its own layout. The page styles are in `src/home/home.css` and `src/protect/protect.css`.

### Theme

The site follows the system theme by default. The toggle in the top bar cycles system, light and dark, and keeps the choice in `localStorage` under `gryph-theme`. An inline script in the head of `index.html` and `protect.html` sets `data-theme` and `data-theme-choice` on the root element before the first paint. `src/shared/theme.ts` holds the same logic for the toggle. Keep the two in step.

The fonts are in `public/fonts/`. Read `public/fonts/LICENSE.txt` before you add or change a font.
