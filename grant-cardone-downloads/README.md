# Grant Cardone Video Downloader

Advanced tool to download all videos from training.grantcardone.com using authentication tokens extracted from your browser session.

## 🚀 Quick Start

### Step 1: Extract Authentication Data

1. Open Chrome and go to https://training.grantcardone.com/library
2. Make sure you're logged in
3. Open Developer Tools (Cmd+Opt+I)
4. Go to the Console tab
5. Copy and paste the contents of `extract_auth.js`
6. Press Enter - this will automatically download `grantcardone_auth.json`

### Step 2: Download Videos

```bash
cd /Users/bird/sources/standalone-projects/grant-cardone-downloads
python3 advanced_extractor.py
```

## 📁 Files

- `advanced_extractor.py` - Main downloader script
- `extract_auth.js` - JavaScript for extracting auth tokens
- `extract_videos.sh` - Alternative extraction method
- `grant_cardone_downloader.py` - Basic downloader (fallback)

## 🔧 Features

- **Stealth Mode**: Mimics browser behavior exactly
- **Auth Token Extraction**: Uses your existing session
- **API Discovery**: Automatically finds video endpoints
- **Batch Downloading**: Downloads all videos in correct order
- **Metadata Preservation**: Keeps titles, descriptions, subtitles
- **Rate Limiting**: Respects server limits with random delays

## 🛡️ Anti-Detection Measures

1. **Realistic Headers**: Uses exact Chrome User-Agent and headers
2. **Random Delays**: Human-like timing between requests
3. **Session Preservation**: Uses your actual browser session
4. **Cookie Management**: Proper cookie handling
5. **Error Recovery**: Automatic retries with exponential backoff

## 📥 Download Quality

- Maximum resolution: 1080p
- Embedded metadata and thumbnails
- Subtitles (if available)
- Original audio quality preserved

## 📂 Output Structure

```
grant-cardone-downloads/
├── Program 1/
│   ├── 001-Video-Title.mp4
│   ├── 002-Another-Video.mp4
│   └── ...
├── Program 2/
│   ├── 001-First-Video.mp4
│   └── ...
└── Individual Videos/
    ├── 001-Standalone-Video.mp4
    └── ...
```

## 🔍 Troubleshooting

### "No video data found"
- Ensure you're logged into Grant Cardone training
- Check that `grantcardone_auth.json` exists and is valid
- Try re-running the auth extraction script

### "Authentication failed"
- Auth tokens may have expired
- Extract fresh auth data and retry
- Ensure you're logged in with same browser session

### "Download failed"
- Check your internet connection
- Try reducing concurrent downloads
- Ensure you have enough disk space

## 📝 Requirements

- Python 3.6+
- yt-dlp (`brew install yt-dlp`)
- Chrome browser
- Valid Grant Cardone training account

## ⚡ Performance Tips

- Use SSD storage for faster downloads
- Ensure stable internet connection
- Close other bandwidth-intensive applications
- Monitor disk space during large downloads

## 🔒 Privacy & Security

- Auth tokens are only used for API calls
- No personal data is stored permanently
- Scripts operate locally on your machine
- Tokens are automatically cleared after use

## 📞 Support

If you encounter issues:
1. Check the troubleshooting section above
2. Ensure all requirements are met
3. Verify your account access to the training platform

---

⚠️ **Disclaimer**: Use responsibly. This tool is for personal backup purposes only. Respect the platform's terms of service.