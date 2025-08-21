
## 1. Project conventions
- Use `bun` commands for running the app.
- Use `bun run dev` instead of `npm run dev`.
- Use `bun add` instead of `npm install`.
- Use `bunx` to auto-install and run packages from npm. It's Bun's equivalent of `npx` or `yarn dlx`.

```bash
bunx cowsay "Hello world!"
```

## 2. EXISTING CODE MODIFICATIONS - BE VERY STRICT

- Any added complexity to existing files needs strong justification
- Always prefer extracting to new controllers/services/ over complicating existing ones
- Question every change: "Does this make the existing code harder to understand?"

## 3. NEW CODE - BE PRAGMATIC

- If it's isolated and works, its's acceptable
- Still flag obvious improvements but don't block progress
- Focus on whether the code is testable and maintainbable
