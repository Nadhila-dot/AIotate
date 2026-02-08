# AIotate README

## What's Included

The AIotate binary is a **single self-contained executable** that includes:
- ✅ Go backend server
- ✅ React frontend (embedded)
- ✅ WebSocket support
- ✅ Desktop webview wrapper
- ✅ All static assets

## What Users Need

### Required
1. **Tectonic LaTeX compiler** - For PDF generation
   - macOS: `brew install tectonic`
   - Windows: `scoop install tectonic`
   - Linux: `sudo apt install tectonic` or `cargo install tectonic`

### Optional
2. **AI API Key** - For worksheet generation
   - Gemini API (free tier available): https://aistudio.google.com/app/apikey
   - OpenRouter API (pay-as-you-go): https://openrouter.ai/keys

## First Run Experience

When users run AIotate for the first time:

### 1. Automatic Setup
The application automatically creates:
```
./
├── set.json                    # Configuration file
├── badger-db/                  # Database
├── storage/
│   ├── bucket/                 # Generated PDFs
│   └── queue_data/             # Queue persistence
├── generated/                  # LaTeX files
│   └── gemini_fixes/           # AI fix attempts
├── zp-database/                # JSON debug exports
│   ├── users/
│   ├── sessions/
│   ├── notebooks/
│   ├── queue/
│   └── styles/
└── logs/                       # Application logs
```

### 2. Tectonic Check
If Tectonic is not installed, the app shows:
```
╔════════════════════════════════════════════════════════════════╗
║                    TECTONIC NOT FOUND                          ║
╚════════════════════════════════════════════════════════════════╝

AIotate requires Tectonic LaTeX compiler to generate PDFs.

Installation instructions:
  [Platform-specific instructions shown here]

After installation, restart AIotate.
```

### 3. Configuration
A default `set.json` is created with empty API keys:
```json
{
  "AI_PROVIDER": "gemini",
  "GEMINI_API_KEY": "",
  "OPENROUTER_API_KEY": "",
  "AI_MAIN_MODEL": "",
  "AI_UTILITY_MODEL": "",
  "MAX_SESSIONS": 2,
  "SHEET_QUEUE_DIR": "./storage/queue_data"
}
```

Users can configure API keys through:
- **Settings page** in the UI (recommended)
- **Manual edit** of `set.json`

### 4. Ready to Use
Once Tectonic is installed and API keys are configured:
- ✅ Create worksheets
- ✅ Organize in notebooks
- ✅ Generate PDFs
- ✅ All features available

## Distribution Package

### Single Binary Distribution
```
aiotate-[platform]/
├── aiotate (or aiotate.exe)
└── README.txt
```

### README.txt Template
```
AIotate - AI-Powered Educational Worksheet Generator
====================================================

REQUIREMENTS:
1. Tectonic LaTeX compiler
   macOS:   brew install tectonic
   Windows: scoop install tectonic
   Linux:   sudo apt install tectonic

2. AI API Key (optional, for generation)
   Get free key: https://aistudio.google.com/app/apikey

FIRST RUN:
1. Run the aiotate executable
2. Application will create necessary directories
3. Configure API key in Settings
4. Start creating worksheets!

SUPPORT:
- Documentation: under construction
- Issues: https://github.com/Nadhila-dot/AIotate/issues
- Website: nadhi.dev

Made with ❤️ by Nadhi.dev
```

## Platform-Specific Notes

### macOS
- **Gatekeeper**: Users may need to right-click → Open on first run
- **Permissions**: May prompt for network access
- **Universal Binary**: Provide separate builds for Intel and Apple Silicon

### Windows
- **SmartScreen**: May show warning on first run
- **Antivirus**: Some may flag the executable (false positive)
- **Dependencies**: No additional DLLs needed

### Linux
- **Permissions**: Make executable: `chmod +x aiotate`
- **Dependencies**: Requires GTK WebKit2 (`libwebkit2gtk-4.0`)
- **Desktop Integration**: Can create .desktop file for menu entry

## File Locations

All user data is stored relative to the executable:
- **Configuration**: `./set.json`
- **Database**: `./badger-db/`
- **Generated PDFs**: `./storage/bucket/`
- **LaTeX files**: `./generated/`

This makes the application **portable** - users can move the entire directory.

## Upgrade Process

To upgrade:
1. Download new binary
2. Replace old binary
3. Keep existing data directories
4. Restart application

All user data (notebooks, settings, PDFs) is preserved.

## Uninstall

To remove AIotate:
1. Delete the binary
2. Delete the data directories (if desired)
3. No registry entries or system files to clean

## Security Notes

- API keys stored in plain text in `set.json`
- Recommend setting file permissions: `chmod 600 set.json`
- Database is local-only (no cloud sync)
- No telemetry or analytics

## Troubleshooting

### "Tectonic not found"
Install Tectonic following platform instructions above.

### "Failed to generate PDF"
1. Check Tectonic is installed: `tectonic --version`
2. Check LaTeX file in `./generated/sheet-[id]/`
3. Try manual compilation: `tectonic file.tex`

### "AI generation failed"
1. Check API key is configured in Settings
2. Verify API key is valid
3. Check internet connection
4. Review logs in `./logs/`

### "Application won't start"
1. Check file permissions
2. **Ensure port 317 is available**
3. Check logs in `./logs/`
4. Try running from terminal to see errors

## Support

For issues or questions:
- GitHub: https://github.com/Nadhila-dot/AIotate
- Email: hello@nadhi.dev
- Website: nadhi.dev
