#!/usr/bin/env python3
"""
STAGE 2: Automated Grant Cardone Video Scraper
Cycles through all courses from stage1 and extracts video URLs
"""

import os
import json
import time
import random
import subprocess
import requests
import webbrowser
from urllib.parse import urljoin, urlparse
import tempfile
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
import re

class AutomatedScraper:
    def __init__(self):
        self.courses = []
        self.videos = []
        self.session = requests.Session()
        self.download_dir = "./grant-cardone-downloads"
        self.chrome_driver = None

        # Configure session to mimic browser
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
            'Accept-Language': 'en-US,en;q=0.5',
            'Accept-Encoding': 'gzip, deflate',
            'DNT': '1',
            'Connection': 'keep-alive',
            'Upgrade-Insecure-Requests': '1'
        })

    def load_courses_from_stage1(self):
        """Load courses from stage1 output"""
        try:
            with open('grantcardone_courses.json', 'r') as f:
                data = json.load(f)
                self.courses = data.get('courses', [])
                print(f"✅ Loaded {len(self.courses)} courses from stage1")
                return True
        except FileNotFoundError:
            print("❌ grantcardone_courses.json not found - run stage1 first")
            return False
        except json.JSONDecodeError:
            print("❌ Invalid JSON in grantcardone_courses.json")
            return False

    def create_stage2_script(self, course_links):
        """Create JavaScript for stage2 extraction"""
        return f"""
// STAGE 2: Video URL Extractor - Auto-generated
(function() {{
    console.log('🎥 STAGE 2: Extracting video URLs...');

    const videos = [];
    const courseLinks = {json.dumps(course_links)};

    // Function to extract videos from current page
    function extractVideosFromPage() {{
        const pageVideos = [];

        // Find video elements
        document.querySelectorAll('video').forEach((video, i) => {{
            const src = video.src || video.querySelector('source')?.src;
            if (src && src.startsWith('http')) {{
                pageVideos.push({{
                    url: src,
                    title: video.title || video.getAttribute('data-title') || `Video ${{i+1}}`,
                    source: 'video-element'
                }});
            }}
        }});

        // Find iframes
        document.querySelectorAll('iframe').forEach((iframe, i) => {{
            if (iframe.src && (iframe.src.includes('vimeo') || iframe.src.includes('youtube') || iframe.src.includes('player'))) {{
                pageVideos.push({{
                    url: iframe.src,
                    title: iframe.title || `Video ${{i+1}}`,
                    source: 'iframe'
                }});
            }}
        }});

        // Look for script tags with video URLs
        document.querySelectorAll('script').forEach((script) => {{
            const text = script.textContent;

            // Look for direct video URLs
            const urlPatterns = [
                /https?:\/\/[^\\s"']+\\.mp4[^\\s"']*/g,
                /https?:\/\/[^\\s"']+\\.m3u8[^\\s"']*/g,
                /https?:\/\/vimeo\.com\/[^\\s"']+/g,
                /https?:\/\/youtu\\.be\/[^\\s"']+/g
            ];

            urlPatterns.forEach(pattern => {{
                const matches = text.match(pattern);
                if (matches) {{
                    matches.forEach(url => {{
                        pageVideos.push({{
                            url: url.trim(),
                            title: `Video ${{pageVideos.length + 1}}`,
                            source: 'script-detection'
                        }});
                    }});
                }}
            }});

            // Look for JSON video data
            try {{
                const jsonMatches = text.match(/\\{{[^{{}}]*"video[^"]*"[^{{}}]*\\}}/g);
                if (jsonMatches) {{
                    jsonMatches.forEach(match => {{
                        try {{
                            const data = JSON.parse(match);
                            const videoUrl = data.video_url || data.stream_url || data.url;
                            const title = data.title || data.name;
                            if (videoUrl && videoUrl.startsWith('http')) {{
                                pageVideos.push({{
                                    url: videoUrl,
                                    title: title || `Video ${{pageVideos.length + 1}}`,
                                    source: 'json-data'
                                }});
                            }}
                        }} catch (e) {{ /* Ignore invalid JSON */ }}
                    }});
                }}
            }} catch (e) {{ /* Ignore */ }}
        }});

        // Remove duplicates
        return pageVideos.filter((video, index, self) =>
            index === self.findIndex((v) => v.url === video.url)
        );
    }}

    // Extract videos from current page
    const currentVideos = extractVideosFromPage();

    console.log(`Found ${{currentVideos.length}} videos on ${{window.location.href}}`);

    // Prepare final data
    const videoData = {{
        videos: currentVideos,
        pageUrl: window.location.href,
        timestamp: new Date().toISOString(),
        courseLinks: courseLinks
    }};

    // Auto-download the data
    const dataStr = JSON.stringify(videoData, null, 2);
    const dataBlob = new Blob([dataStr], {{type: 'application/json'}});
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'stage2_videos.json';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    console.log('✅ Stage 2 data downloaded as stage2_videos.json');
    console.log(JSON.stringify(currentVideos, null, 2));

    return {{
        videos: currentVideos,
        totalVideos: currentVideos.length,
        pageUrl: window.location.href
    }};
}})();
"""

    def create_automation_script(self):
        """Create a comprehensive automation script for browser"""
        automation_js = """
// Grant Cardone Complete Automation Script
(function() {
    const automationData = {
        stage: 'complete_automation',
        courses: [],
        videos: [],
        timestamp: new Date().toISOString()
    };

    // Function to extract course info from current page
    function extractCourseInfo() {
        const courses = [];

        // Find all "Start Now" buttons
        const buttons = Array.from(document.querySelectorAll('button, a')).filter(el => {
            const text = (el.textContent || '').toLowerCase();
            return text.includes('start now') || text.includes('start') || text.includes('begin');
        });

        buttons.forEach((button, index) => {
            const container = button.closest('.course, .lesson, .program, [class*="course"], [class*="lesson"]');

            let title = '';
            let link = '';

            if (container) {
                const titleEl = container.querySelector('h1, h2, h3, h4, .title, .course-title');
                title = titleEl?.textContent?.trim() || `Course ${index + 1}`;

                const linkEl = container.querySelector('a[href]');
                link = linkEl?.href || '';
            }

            if (!title) title = button.textContent.trim() || `Course ${index + 1}`;

            courses.push({
                id: index + 1,
                title: title,
                link: link,
                buttonText: button.textContent.trim()
            });
        });

        return courses;
    }

    // Function to extract videos from current page
    function extractVideos() {
        const videos = [];

        // Video elements
        document.querySelectorAll('video').forEach((video, i) => {
            const src = video.src || video.querySelector('source')?.src;
            if (src && src.startsWith('http')) {
                videos.push({
                    url: src,
                    title: video.title || `Video ${i+1}`,
                    source: 'video-element'
                });
            }
        });

        // Iframes
        document.querySelectorAll('iframe').forEach((iframe, i) => {
            if (iframe.src && (iframe.src.includes('vimeo') || iframe.src.includes('youtube'))) {
                videos.push({
                    url: iframe.src,
                    title: iframe.title || `Video ${i+1}`,
                    source: 'iframe'
                });
            }
        });

        // Script detection
        document.querySelectorAll('script').forEach((script) => {
            const text = script.textContent;

            const urlPatterns = [
                /https?:\/\/[^\s"']+\.mp4[^\s"']*/g,
                /https?:\/\/[^\s"']+\.m3u8[^\s"']*/g
            ];

            urlPatterns.forEach(pattern => {
                const matches = text.match(pattern);
                if (matches) {
                    matches.forEach(url => {
                        videos.push({
                            url: url.trim(),
                            title: `Video ${videos.length + 1}`,
                            source: 'script-detection'
                        });
                    });
                }
            });
        });

        return videos.filter((video, index, self) =>
            index === self.findIndex((v) => v.url === video.url)
        );
    }

    // Extract everything
    automationData.courses = extractCourseInfo();
    automationData.videos = extractVideos();

    console.log('🎥 AUTOMATION COMPLETE');
    console.log(`📚 Courses: ${automationData.courses.length}`);
    console.log(`🎬 Videos: ${automationData.videos.length}`);

    // Auto-download
    const dataStr = JSON.stringify(automationData, null, 2);
    const dataBlob = new Blob([dataStr], {type: 'application/json'});
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'grantcardone_complete.json';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    return automationData;
})();
"""

        with open('complete_automation.js', 'w') as f:
            f.write(automation_js)

        return 'complete_automation.js'

    def run_browser_automation(self):
        """Run complete browser automation"""
        print("🤖 Starting complete browser automation...")

        script_path = self.create_automation_script()

        print("📜 AUTOMATION SCRIPT READY")
        print("1. Open Chrome → https://training.grantcardone.com/library")
        print("2. Open DevTools (Cmd+Opt+I) → Console")
        print("3. Copy contents of complete_automation.js")
        print("4. Paste in console and press Enter")
        print("5. JSON will auto-download as grantcardone_complete.json")
        print("6. Then run: python3 stage2_automated_scraper.py")

        return script_path

    def setup_aria2c(self):
        """Setup aria2c configuration for multi-threaded downloads"""
        aria2_config = f"""
# Grant Cardone Video Downloader Configuration
max-concurrent-downloads=16
split=16
min-split-size=1M
max-connection-per-server=16
piece-length=1M
continue=true
max-tries=5
retry-wait=30
timeout=600
connect-timeout=60
allow-overwrite=true
file-allocation=trunc
dir={os.path.abspath(self.download_dir)}
log-level=notice
log=aria2c.log
console-log-level=notice
summary-interval=60
show-console-readout=true
enable-rpc=false
"""

        with open('aria2.conf', 'w') as f:
            f.write(aria2_config)

        print("✅ aria2c configuration created: aria2.conf")
        return 'aria2.conf'

    def download_with_aria2c(self, videos):
        """Download videos using aria2c with multi-threading"""
        if not videos:
            print("❌ No videos to download")
            return

        print(f"🚀 Starting aria2c download of {len(videos)} videos...")

        # Setup aria2c
        self.setup_aria2c()
        os.makedirs(self.download_dir, exist_ok=True)

        # Create aria2c input file
        input_file = 'aria2_downloads.txt'
        with open(input_file, 'w') as f:
            for i, video in enumerate(videos, 1):
                title = video.get('title', f'Video {i}')
                safe_title = re.sub(r'[^\w\s-]', '', title).strip()
                safe_title = re.sub(r'[-\s]+', '-', safe_title)
                filename = f"{i:03d}-{safe_title}"

                # Create aria2c line format
                url = video['url']
                out_path = os.path.join(self.download_dir, f"{filename}.%(ext)s")
                f.write(f"{url}\n")
                f.write(f"  out={os.path.basename(out_path)}\n")
                f.write("\n")

        # Run aria2c
        cmd = [
            'aria2c',
            '--conf-path=aria2.conf',
            '--input-file=' + input_file,
            '--show-console-readout',
            '--summary-interval=30'
        ]

        try:
            print("🔥 Starting aria2c with 16 parallel connections...")
            result = subprocess.run(cmd, cwd=os.getcwd())

            if result.returncode == 0:
                print("✅ All downloads completed successfully!")
            else:
                print(f"⚠️ aria2c completed with exit code: {result.returncode}")

        except subprocess.CalledProcessError as e:
            print(f"❌ aria2c error: {e}")
        except FileNotFoundError:
            print("❌ aria2c not found. Install with: brew install aria2")

        print(f"📁 Videos saved to: {os.path.abspath(self.download_dir)}")

    def load_complete_data(self):
        """Load data from complete automation"""
        try:
            with open('grantcardone_complete.json', 'r') as f:
                data = json.load(f)
                videos = data.get('videos', [])
                print(f"✅ Loaded {len(videos)} videos from automation")
                return videos
        except FileNotFoundError:
            print("❌ grantcardone_complete.json not found")
            return []
        except json.JSONDecodeError:
            print("❌ Invalid JSON in grantcardone_complete.json")
            return []

    def run(self):
        """Main execution"""
        print("🎥 STAGE 2: Automated Grant Cardone Video Scraper")
        print("=" * 60)

        # Try to load complete automation data first
        videos = self.load_complete_data()

        if not videos:
            # No automation data, show instructions
            script_path = self.run_browser_automation()
            return

        # If we have videos, download them
        if videos:
            print(f"\n🎬 Found {len(videos)} videos! Starting downloads...")
            self.download_with_aria2c(videos)
        else:
            print("❌ No videos found in automation data")

if __name__ == "__main__":
    scraper = AutomatedScraper()
    scraper.run()