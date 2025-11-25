#!/usr/bin/env python3
"""
Interactive Grant Cardone Video Downloader
Provides step-by-step guidance for manual extraction
"""

import os
import json
import time
import subprocess
import re

class InteractiveExtractor:
    def __init__(self):
        self.download_dir = "./grant-cardone-downloads"

    def show_step_by_step(self):
        """Show step-by-step instructions"""
        print("🎥 INTERACTIVE GRANT CARDONE VIDEO EXTRACTOR")
        print("=" * 60)

        print("\n📋 STEP-BY-STEP INSTRUCTIONS:")
        print("-" * 40)

        print("\n🔍 STEP 1: Navigate to Video Content")
        print("1. Open Chrome → https://training.grantcardone.com/library")
        print("2. Click on your programs/courses to see actual video lessons")
        print("3. Navigate to pages where videos play")

        print("\n🔍 STEP 2: Extract Video URLs")
        print("1. Open Developer Tools (Cmd+Opt+I) → Console")
        print("2. Paste this JavaScript and press Enter:")

        js_code = '''
// Simple Video URL Extractor
(function() {
    const videos = [];

    // Look for video elements
    document.querySelectorAll('video').forEach((video, i) => {
        if (video.src) {
            videos.push({url: video.src, title: video.title || `Video ${i+1}`});
        }
        const source = video.querySelector('source');
        if (source && source.src) {
            videos.push({url: source.src, title: video.title || `Video ${i+1}`});
        }
    });

    // Look for iframes
    document.querySelectorAll('iframe').forEach((iframe, i) => {
        if (iframe.src && (iframe.src.includes('vimeo') || iframe.src.includes('youtube'))) {
            videos.push({url: iframe.src, title: iframe.title || `Video ${i+1}`});
        }
    });

    // Look for lesson/course links
    document.querySelectorAll('a[href*="lesson"], a[href*="video"], .lesson-item a, .video-item a').forEach((link, i) => {
        const title = link.textContent.trim() || link.title || `Lesson ${i+1}`;
        videos.push({url: link.href, title: title, type: 'lesson-link'});
    });

    console.log('Found', videos.length, 'videos/lessons:');
    videos.forEach((v, i) => {
        console.log(`${i+1}. ${v.title}`);
        console.log(`   URL: ${v.url}`);
    });

    // Return as downloadable JSON
    const data = JSON.stringify(videos, null, 2);
    console.log('\\nCOPY THIS DATA:');
    console.log(data);

    return videos;
})();
'''

        print(js_code)

        print("\n🔍 STEP 3: Collect Video Data")
        print("1. Copy the JSON output from the console")
        print("2. Save it as 'found_videos.json' in this directory")

        print("\n🔍 STEP 4: Download Videos")
        print("1. Run: python3 interactive_extractor.py")
        print("2. The script will download all found videos")

    def download_collected_videos(self):
        """Download videos from collected data"""
        try:
            # Find which video file exists
            video_file = None
            if os.path.exists('grantcardone_videos.json'):
                video_file = 'grantcardone_videos.json'
            elif os.path.exists('found_videos.json'):
                video_file = 'found_videos.json'

            if not video_file:
                print("❌ No video data file found")
                return

            with open(video_file, 'r') as f:
                data = json.load(f)
                videos = data.get('videos', data)  # Handle both formats

            if not videos:
                print("❌ No videos found in found_videos.json")
                return

            print(f"\n🎬 Found {len(videos)} videos to download!")

            os.makedirs(self.download_dir, exist_ok=True)

            successful = 0
            failed = 0

            for i, video in enumerate(videos, 1):
                title = video.get('title', f'Video {i}')
                url = video['url']

                print(f"\n[{i}/{len(videos)}] {title}")
                print(f"URL: {url}")

                # Create safe filename
                safe_title = re.sub(r'[^\w\s-]', '', title).strip()
                safe_title = re.sub(r'[-\s]+', '-', safe_title)
                filename = f"{i:03d}-{safe_title}"

                output_path = os.path.join(self.download_dir, f"{filename}.%(ext)s")

                # Handle lesson links differently
                if video.get('type') == 'lesson-link':
                    print(f"📋 Lesson link - will need manual access: {url}")
                    # You could add code to navigate to lesson pages here
                    continue

                # yt-dlp command
                cmd = [
                    'yt-dlp',
                    '--no-warnings',
                    '--embed-metadata',
                    '--output', output_path,
                    '--format', 'best[height<=1080]',
                    '--retries', '3',
                    url
                ]

                try:
                    print(f"📥 Downloading...")
                    result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)

                    if result.returncode == 0:
                        print(f"✅ Success: {title}")
                        successful += 1
                    else:
                        print(f"❌ Failed: {title}")
                        if result.stderr:
                            print(f"   Error: {result.stderr[:200]}")
                        failed += 1

                except subprocess.TimeoutExpired:
                    print(f"⏰ Timeout: {title}")
                    failed += 1
                except Exception as e:
                    print(f"❌ Error: {title} - {e}")
                    failed += 1

                # Delay between downloads
                time.sleep(2)

            print(f"\n📊 DOWNLOAD SUMMARY:")
            print(f"✅ Successful: {successful}")
            print(f"❌ Failed: {failed}")
            print(f"📁 Videos saved to: {os.path.abspath(self.download_dir)}")

        except FileNotFoundError:
            print("\n⚠️  found_videos.json not found")
            print("Please follow the extraction steps first!")
        except json.JSONDecodeError:
            print("❌ Invalid JSON in found_videos.json")

    def run(self):
        """Main execution"""
        # Check if we have video data (try both possible filenames)
        video_file = None
        if os.path.exists('grantcardone_videos.json'):
            video_file = 'grantcardone_videos.json'
        elif os.path.exists('found_videos.json'):
            video_file = 'found_videos.json'

        if video_file:
            print("🎬 Found video data! Starting downloads...")
            self.download_collected_videos()
        else:
            print("📋 No video data found. Showing extraction instructions...")
            self.show_step_by_step()

            print(f"\n📁 Save your extracted video data to: {os.path.abspath('found_videos.json')}")
            print("Then run this script again to download all videos!")

if __name__ == "__main__":
    extractor = InteractiveExtractor()
    extractor.run()