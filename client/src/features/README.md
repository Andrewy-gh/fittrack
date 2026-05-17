# Features

Feature folders group code by product behavior instead of by file type.

## Structure

- `routes/` keeps TanStack Router files, loaders, search params, and route wiring.
- `features/<area>/` owns UI and local helpers for one product area.
- `components/` stays for shared app shell, generic controls, and design-system UI.
- `lib/` stays for cross-feature data access, persistence, and pure product logic.

## Import Rule

Routes can import feature modules. Features can import shared UI and `lib` modules.
Avoid importing route-private files from a feature; move the reusable code into the
feature instead.

Shared behavior used by multiple product areas can become its own feature, like
`metric-charts`, when it has meaningful behavior beyond basic UI.
