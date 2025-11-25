#!/usr/bin/env python3
"""
Grant Cardone Video Downloader
Downloads all videos from training.grantcardone.com in correct order
"""

import os
import sys
import json
import time
import subprocess
import requests
from urllib.parse import urljoin, urlparse
import re

class GrantCardoneDownloader:
    def __init__(self, download_dir="./grant-cardone-videos"):
        self.download_dir = download_dir
        self.session = requests.Session()
        self.chrome_debug_url = "http://localhost:9222"
        self.videos = []

    def setup_download_directory(self):
        """Create download directory if it doesn't exist"""
        if not os.path.exists(self.download_dir):
            os.makedirs(self.download_dir)
            print(f"Created download directory: {self.download_dir}")

    def check_chrome_debugging(self):
        """Check if Chrome is running with remote debugging"""
        try:
            response = requests.get(f"{self.chrome_debug_url}/json")
            if response.status_code == 200:
                tabs = response.json()
                for tab in tabs:
                    if "training.grantcardone.com" in tab.get("url", ""):
                        return tab
                print("Chrome debugging is running but Grant Cardone tab not found")
                print("Available tabs:")
                for tab in tabs:
                    print(f"  - {tab.get('url', 'Unknown URL')}")
                return None
            else:
                print("Chrome remote debugging not responding")
                return None
        except Exception as e:
            print(f"Error connecting to Chrome debug port: {e}")
            print("Please start Chrome with remote debugging:")
            print("  /Applications/Google\\ Chrome.app/Contents/MacOS/Google\\ Chrome --remote-debugging-port=9222")
            return None

    def extract_videos_from_tab(self, tab):
        """Extract video information from Chrome tab"""
        try:
            websocket_url = tab.get("webSocketDebuggerUrl")
            if not websocket_url:
                print("No WebSocket URL found for tab")
                return []

            # For now, we'll use a simpler approach by getting the page content
            # and extracting video URLs manually
            return self.extract_videos_manually()

        except Exception as e:
            print(f"Error extracting videos from tab: {e}")
            return []

    def extract_videos_manually(self):
        """
        Manually extract video information using browser developer tools
        This method requires user to run some JavaScript in the browser console
        """
        print("\n" + "="*60)
        print("MANUAL VIDEO EXTRACTION REQUIRED")
        print("="*60)
        print("\nPlease follow these steps:")
        print("1. Open Chrome Developer Tools (Cmd+Opt+I)")
        print("2. Go to the Console tab")
        print("3. Paste and run this JavaScript code:")
        print("\n" + "-"*50)

        js_code = """
// Extract all video information from the page
(function() {
    const videos = [];

    // Look for video elements
    const videoElements = document.querySelectorAll('video');
    videoElements.forEach((video, index) => {
        videos.push({
            type: 'video-element',
            src: video.src,
            poster: video.poster,
            title: video.title || `Video ${index + 1}`,
            index: index
        });
    });

    // Look for iframe elements (common for video players)
    const iframes = document.querySelectorAll('iframe');
    iframes.forEach((iframe, index) => {
        if (iframe.src.includes('vimeo') || iframe.src.includes('youtube') || iframe.src.includes('wistia')) {
            videos.push({
                type: 'iframe',
                src: iframe.src,
                title: iframe.title || `Video ${videos.length + 1}`,
                index: videos.length
            });
        }
    });

    // Look for links that might be videos
    const links = document.querySelectorAll('a[href*="video"], a[href*="vimeo"], a[href*="youtube"]');
    links.forEach((link, index) => {
        videos.push({
            type: 'link',
            src: link.href,
            title: link.textContent.trim() || `Video ${videos.length + 1}`,
            index: videos.length
        });
    });

    // Look for course/program structure
    const courseItems = document.querySelectorAll('.course-item, .lesson-item, .video-item, [class*="lesson"], [class*="video"], [class*="course"]');
    courseItems.forEach((item, index) => {
        const link = item.querySelector('a');
        const title = item.querySelector('.title, .lesson-title, .video-title') || item;

        if (link || title) {
            videos.push({
                type: 'course-item',
                src: link ? link.href : '',
                title: title.textContent.trim() || `Course Item ${index + 1}`,
                index: videos.length,
                element: item.className
            });
        }
    });

    // Try to find JSON data in script tags
    const scripts = document.querySelectorAll('script');
    scripts.forEach(script => {
        try {
            const text = script.textContent;
            if (text.includes('video') && (text.includes('"url"') || text.includes('"src"'))) {
                // Try to extract JSON data
                const jsonMatch = text.match(/\{[\s\S]*\}/);
                if (jsonMatch) {
                    try {
                        const jsonData = JSON.parse(jsonMatch[0]);
                        if (jsonData.videos || jsonData.video || jsonData.url) {
                            videos.push({
                                type: 'json-data',
                                data: jsonData,
                                title: 'JSON Video Data',
                                index: videos.length
                            });
                        }
                    } catch (e) {
                        // Ignore JSON parse errors
                    }
                }
            }
        } catch (e) {
            // Ignore script processing errors
        }
    });

    console.log('Found ' + videos.length + ' video elements:', videos);
    console.log('Copy this JSON for the downloader:');
    console.log(JSON.stringify(videos, null, 2));

    return videos;
})();
"""

        print(js_code)
        print("-"*50)
        print("\n4. Copy the JSON output from the console")
        print("5. Create a file named 'video_data.json' in this directory")
        print("6. Paste the JSON data into that file")
        print("7. Press Enter to continue...")

        input()

        # Load the manually extracted video data
        try:
            with open('video_data.json', 'r') as f:
                video_data = json.load(f)
                return self.process_video_data(video_data)
        except FileNotFoundError:
            print("Error: video_data.json file not found")
            return []
        except json.JSONDecodeError:
            print("Error: Invalid JSON in video_data.json")
            return []

    def process_video_data(self, video_data):
        """Process raw video data into downloadable URLs"""
        processed_videos = []

        for item in video_data:
            if isinstance(item, dict):
                title = item.get('title', f'Video {len(processed_videos) + 1}')
                src = item.get('src', '')

                # Clean up filename
                safe_title = re.sub(r'[^\w\s-]', '', title).strip()
                safe_title = re.sub(r'[-\s]+', '-', safe_title)

                if src:
                    processed_videos.append({
                        'title': title,
                        'filename': f"{len(processed_videos) + 1:03d}-{safe_title}",
                        'url': src,
                        'type': item.get('type', 'unknown')
                    })

        return processed_videos

    def download_video(self, video_info):
        """Download a single video using yt-dlp"""
        url = video_info['url']
        filename = video_info['filename']
        title = video_info['title']

        print(f"\nDownloading: {title}")
        print(f"URL: {url}")
        print(f"Filename: {filename}")

        # yt-dlp command
        cmd = [
            'yt-dlp',
            '--no-warnings',
            '--embed-metadata',
            '--embed-thumbnail',
            '--write-subtitles',
            '--write-auto-subs',
            '--sub-langs', 'en',
            '--output', f"{self.download_dir}/{filename}.%(ext)s",
            '--format', 'best[height<=1080]',
            url
        ]

        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            print(f"✅ Successfully downloaded: {title}")
            return True
        except subprocess.CalledProcessError as e:
            print(f"❌ Failed to download {title}: {e}")
            print(f"Error output: {e.stderr}")
            return False

    def download_all_videos(self, videos):
        """Download all videos in order"""
        if not videos:
            print("No videos to download")
            return

        print(f"Starting download of {len(videos)} videos...")
        print(f"Download directory: {os.path.abspath(self.download_dir)}")

        successful_downloads = 0
        failed_downloads = 0

        for i, video in enumerate(videos, 1):
            print(f"\n{'='*60}")
            print(f"Video {i}/{len(videos)}")
            print(f"{'='*60}")

            if self.download_video(video):
                successful_downloads += 1
            else:
                failed_downloads += 1

            # Small delay between downloads to be respectful
            time.sleep(1)

        print(f"\n{'='*60}")
        print("DOWNLOAD SUMMARY")
        print(f"{'='*60}")
        print(f"Total videos: {len(videos)}")
        print(f"Successful: {successful_downloads}")
        print(f"Failed: {failed_downloads}")
        print(f"Download directory: {os.path.abspath(self.download_dir)}")

    def run(self):
        """Main execution method"""
        print("🎥 Grant Cardone Video Downloader")
        print("="*50)

        self.setup_download_directory()

        # Try to get video data from browser
        videos = []

        # Check if Chrome remote debugging is available
        tab = self.check_chrome_debugging()
        if tab:
            print("Chrome remote debugging detected")
            videos = self.extract_videos_from_tab(tab)
        else:
            print("Falling back to manual extraction")
            videos = self.extract_videos_manually()

        if videos:
            print(f"\nFound {len(videos)} videos to download")
            self.download_all_videos(videos)
        else:
            print("No videos found. Please check the extraction process.")

if __name__ == "__main__":
    downloader = GrantCardoneDownloader()
    downloader.run()