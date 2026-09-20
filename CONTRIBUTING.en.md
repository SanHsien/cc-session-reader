# Contributing Guide

[繁體中文](CONTRIBUTING.md) | [English](CONTRIBUTING.en.md)

Thank you for your interest in this project! This repository is a maintained fork tailored for Windows and multi-agent AI workflows (Codex, Claude, Cursor, Antigravity, Hermes).

## Principles & Guidelines

1. **Target SanHsien's Repository Only**:
   - PRs and pushes must always target `SanHsien/cc-session-reader` (`origin`).
   - **Do not open PRs or push commits to the upstream repository.**
2. **Windows-First**:
   - Streamlined specifically for Windows 11 + PowerShell environments.
3. **Verification Gate**:
   - Run the Windows verification gate before committing:
     ```powershell
     powershell -NoProfile -File tools/dev_check.ps1
     ```
   - Ensure the check outputs `WINDOWS DEV CHECK GREEN`.
4. **Bilingual Public Documentation**:
   - Keep Traditional Chinese and English versions of upstream public Markdown synchronized. Fork-local governance files may remain Traditional Chinese only.
