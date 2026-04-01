# cookey

CLI tool to capture authenticated browser sessions from iPhone as
Playwright-compatible storageState JSON.

## Install

- macOS: https://github.com/Lakr233/Cookey/releases/latest/download/cookey-macOS.zip
- Linux x86_64: https://github.com/Lakr233/Cookey/releases/latest/download/cookey-linux-x86_64.zip
- Linux aarch64: https://github.com/Lakr233/Cookey/releases/latest/download/cookey-linux-aarch64.zip
- This file: https://github.com/Lakr233/Cookey/releases/latest/download/skills.md

## Commands

- `cookey login <url>` — start session capture, shows QR code
  - `cookey login <url> --update` — refresh an existing session instead of starting from scratch
  - `cookey login <url> --no-detach` — wait inline instead of detaching to a background daemon
  - `cookey login <url> --json` — output machine-readable JSON (request ID, status, etc.)
- `cookey status [rid]` — check if session arrived
- `cookey export <rid>` — print storageState.json to stdout

## Usage

1. `cookey login <target_url>` → QR code appears
2. User scans QR with Cookey iPhone app and logs in
3. `cookey export <rid> > storageState.json`
4. Pass storageState to Playwright or browser automation
5. `cookey login <target_url> --update` → refresh an expired session using the existing cookies as a seed
