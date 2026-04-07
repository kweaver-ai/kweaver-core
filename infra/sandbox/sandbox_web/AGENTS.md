# Repository Guidelines

## Project Structure & Module Organization
`sandbox_web/` is the React frontend for the sandbox platform. Application code lives in `src/`, with API clients under `src/apis/`, reusable layout and UI pieces in `src/components/`, route pages in `src/pages/`, shared types in `src/types/`, and styling in `src/styles/`. Static assets belong in `public/`. Build output is generated into `dist/` and should not be edited manually.

## Build, Test, and Development Commands
Use `npm run dev` to start the Rsbuild development server. Run `npm run build` to create a production bundle, and `npm run preview` to inspect the built app locally. Use `npm run lint` to check TypeScript and React source with ESLint. Run `npm run format` to apply Prettier formatting to `src/**/*.{ts,tsx,less}`.

## Coding Style & Naming Conventions
Write TypeScript with 2-space indentation unless the file already differs. Name React components in `PascalCase` and keep one primary component per file when practical, for example `Sidebar.tsx`. Hooks should use the `useX.ts` pattern, such as `useSessions.ts`. Keep API modules grouped by resource with `api.ts`, `types.ts`, and `index.ts` files. Prefer shared constants in `src/constants/` over inline strings.

## Testing Guidelines
This package currently defines lint and format scripts, but no first-class automated test command in `package.json`. At minimum, run `npm run lint` and `npm run build` before submitting changes. If you add tests later, place them under `tests/` or next to the feature using the project’s existing structure, and document the command in `package.json`.

## Commit & Pull Request Guidelines
Follow the repository’s Conventional Commit style, such as `feat:`, `fix(scope):`, and `docs(scope):`. Keep frontend commits focused on one user-visible change or refactor. Pull requests should describe the UI impact, list any API dependencies, and include screenshots or short recordings for page or layout changes.
