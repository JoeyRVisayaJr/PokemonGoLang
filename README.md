# Pokémon Battle Simulator

A lightweight web app that simulates turn-based Pokémon-style battles using a Go backend and a responsive HTML/CSS/JS frontend.

## Project description

Pokémon Battle Simulator is a compact educational and entertainment project that combines server-side API aggregation with a tactile client-side battle UI. Players load Pokémon from the PokéAPI, select moves, and engage in battles augmented by quick-time-event (QTE) math challenges, XP progression, and upgrade choices.

This app is useful because it demonstrates practical web patterns (server-side fetching, client-side state persistence, non-blocking UI), while also adding an interactive, low-friction learning layer through QTE math tasks. It’s ideal for demos, classroom exercises, and hobby projects that want a small, extensible battle simulation without large frameworks.

## How to upload to GitHub

Prerequisites:
- `git` installed and configured with your name/email
- optionally: GitHub CLI `gh` for automated repo creation and pushing

Steps (recommended):

1. Initialize local repo and commit (from project root):

```powershell
git init
git add -A
git commit -m "Initial commit"
```

2a. Create and push remote using GitHub CLI (recommended):

```powershell
# replace REPO-NAME or omit to infer
gh repo create YOUR_USERNAME/REPO-NAME --public --source=. --remote=origin --push
```

2b. Or create the repo on github.com manually, then push:

```powershell
git remote add origin https://github.com/YOUR_USERNAME/REPO-NAME.git
git branch -M main
git push -u origin main
```

## Quick local run

```powershell
go run main.go
# then open http://localhost:8081
```

## Notes
- The repository already includes a `.gitignore` to avoid committing build artifacts.
- If you want, run the included `publish.ps1` script to automate the local setup and optionally create a GitHub repo with `gh`.
