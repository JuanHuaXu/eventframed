## 2026-08-28 - Installed OpenClaw version is not an npm release

- Category: error
- Context: Installing the adapter with an exact `openclaw@2026.6.2` dependency failed.
- Evidence: npm returned `ETARGET`; the local Homebrew installation reports 2026.6.2 while npm publishes a different release line.
- Lesson: Declare OpenClaw as a compatible peer dependency and use a published package only for development-time SDK type checking.
- Scope: Project-local.

## 2026-08-28 - OpenClaw helper is exported from a subpath

- Category: error
- Context: The adapter imported `definePluginEntry` from the broad plugin SDK entrypoint.
- Evidence: TypeScript reported no such export for the published development package; its export map provides `openclaw/plugin-sdk/plugin-entry`.
- Lesson: Import plugin registration helpers from their documented SDK subpath and compile against the minimum and development versions before release.
- Scope: Project-local.
