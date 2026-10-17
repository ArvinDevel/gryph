// An SVG arrow sits on the optical centre of the text. A text arrow sits low in most fonts.
export function Arrow({ back }: { back?: boolean }) {
  return (
    <svg
      className="arrow"
      viewBox="0 0 16 16"
      aria-hidden="true"
      style={back ? { transform: 'scaleX(-1)' } : undefined}
    >
      <path
        d="M2.5 8h11M9 3.5 13.5 8 9 12.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="square"
        strokeLinejoin="miter"
      />
    </svg>
  )
}

// The pixel GitHub mark from the Cartridge design system, on a 24 by 24 grid.
const GITHUB_PIXEL =
  'M5 2h4v2H7v2H5V2Zm0 10H3V6h2v6Zm2 2H5v-2h2v2Zm2 2v-2H7v2H3v-2H1v2h2v2h4v4h2v-4h2v-2H9Zm0 0v2H7v-2h2Zm6-12v2H9V4h6Zm4 2h-2V4h-2V2h4v4Zm0 6V6h2v6h-2Zm-2 2v-2h2v2h-2Zm-2 2v-2h2v2h-2Zm0 2h-2v-2h2v2Zm0 0h2v4h-2v-4Z'

export function GitHubMark() {
  return (
    <svg className="gh" viewBox="0 0 24 24" aria-hidden="true" shapeRendering="crispEdges">
      <path fill="currentColor" d={GITHUB_PIXEL} />
    </svg>
  )
}

export function Check() {
  return (
    <svg className="check" viewBox="0 0 10 10" aria-hidden="true">
      <path d="M1.5 5.5 4 8l4.5-6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="square" />
    </svg>
  )
}

export function Cross() {
  return (
    <svg className="check" viewBox="0 0 10 10" aria-hidden="true">
      <path d="M2 2l6 6M8 2 2 8" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="square" />
    </svg>
  )
}

// Pixel icons for the theme toggle, on a 16 by 16 grid.
const THEME_PIXEL = {
  system: 'M1 2h14v9H1V2Zm2 2v5h10V4H3Zm3 8h4v2H6v-2Zm-2 2h8v1H4v-1Z',
  light: 'M5 5h6v6H5V5Zm2-4h2v2H7V1Zm0 12h2v2H7v-2ZM1 7h2v2H1V7Zm12 0h2v2h-2V7ZM3 3h2v2H3V3Zm8 0h2v2h-2V3ZM3 11h2v2H3v-2Zm8 0h2v2h-2v-2Z',
  dark: 'M5 1h6v2H8v2H7v6h1v2h3v2H5v-2H3v-2H2V5h1V3h2V1Z',
}

// All three icons render. CSS shows the one for the current choice, also before the page hydrates.
export function ThemeIcons() {
  return (
    <>
      {Object.entries(THEME_PIXEL).map(([choice, d]) => (
        <svg key={choice} className={`gh theme-${choice}`} viewBox="0 0 16 16" aria-hidden="true" shapeRendering="crispEdges">
          <path fill="currentColor" fillRule="evenodd" d={d} />
        </svg>
      ))}
    </>
  )
}
