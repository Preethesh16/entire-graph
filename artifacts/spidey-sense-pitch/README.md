# Spidey Sense pitch deck

- `Spidey-Sense-Pitch-Deck.pptx` — editable 16:9 PowerPoint deck
- `rendered/Spidey-Sense-Pitch-Deck.pdf` — portable presentation copy
- `deck-preview.png` — ten-slide contact sheet
- `build-deck.cjs` — deterministic PptxGenJS source
- `assets/` — clean production screenshots used by the deck

The narrative covers the coordination problem, browser-based team onboarding,
dynamic assignment, Graph Space, confidence-labelled evidence, the real
`HardCoders_` repository analysis, privacy architecture, and the live demo flow.

Rebuild with the workspace-bundled Node dependencies:

```sh
NODE_PATH=/home/infinity/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules \
  node artifacts/spidey-sense-pitch/build-deck.cjs
```
