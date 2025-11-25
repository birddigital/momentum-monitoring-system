#!/usr/bin/env python3
"""
Grant Cardone Video Downloader - Master Script
Complete 2-stage automation with aria2c multi-threaded downloads
"""

import os
import json
import time
import subprocess
import webbrowser
import re
from pathlib import Path

class GrantCardoneMaster:
    def __init__(self):
        self.base_dir = Path(__file__).parent
        self.download_dir = self.base_dir / "grant-cardone-downloads"
        self.download_dir.mkdir(exist_ok=True)

    def check_aria2c(self):
        """Check if aria2c is installed"""
        try:
            result = subprocess.run(['aria2c', '--version'], capture_output=True, text=True)
            if result.returncode == 0:
                print("✅ aria2c detected and ready!")
                return True
        except FileNotFoundError:
            pass

        print("❌ aria2c not found. Installing...")
        try:
            subprocess.run(['brew', 'install', 'aria2'], check=True)
            print("✅ aria2c installed successfully!")
            return True
        except subprocess.CalledProcessError:
            print("❌ Failed to install aria2c. Please run: brew install aria2")
            return False

    def create_aria2_config(self):
        """Create optimized aria2c configuration"""
        config = f"""# Grant Cardone Ultra-Fast Downloader Configuration
max-concurrent-downloads=16
split=16
min-split-size=1M
max-connection-per-server=16
piece-length=1M
continue=true
max-tries=5
retry-wait=30
timeout=600
connect-timeimeout=60
allow-overwrite=true
file-allocation=trunc
dir={self.download_dir.absolute()}
log-level=notice
log=aria2c.log
console-log-level=notice
summary-interval=60
show-console-readout=true
enable-rpc=false
disk-cache=64M
max-file-not-found=5
max-file-not-found=5
parameterized-uri=true
enable-http-keep-alive=true
enable-http-pipelining=true
check-certificate=false
http-no-cache=true
"""

        config_path = self.base_dir / "aria2.conf"
        with open(config_path, 'w') as f:
            f.write(config)

        print(f"✅ aria2c config created: {config_path}")
        return config_path

    def stage1_instructions(self):
        """Show Stage 1 instructions"""
        print("\n🎯 STAGE 1: Extract Course Links")
        print("=" * 50)
        print("1. Open Chrome → https://training.grantcardone.com/library")
        print("2. Open Developer Tools (Cmd+Opt+I) → Console")
        print("3. Copy and paste this JavaScript:")
        print("-" * 30)

        with open(self.base_dir / "stage1_library_scraper.js", 'r') as f:
            js_content = f.read()
            print(js_content[:800] + "...") # Show first part

        print("\n" + "-" * 30)
        print("4. Press Enter")
        print("5. File will auto-download as: grantcardone_courses.json")
        print("6. Return here and press Enter to continue")

        input("Press Enter when you have grantcardone_courses.json...")

    def check_stage1_complete(self):
        """Check if stage1 data exists"""
        courses_file = self.base_dir / "grantcardone_courses.json"
        if courses_file.exists():
            try:
                with open(courses_file, 'r') as f:
                    data = json.load(f)
                    courses = data.get('courses', [])
                    if courses:
                        print(f"✅ Stage 1 complete: {len(courses)} courses found")
                        return courses
            except json.JSONDecodeError:
                pass

        print("❌ Stage 1 data not found or invalid")
        return None

    def stage2_instructions(self):
        """Show Stage 2 instructions"""
        print("\n🎯 STAGE 2: Extract Video URLs")
        print("=" * 50)
        print("Now we'll extract video URLs from each course page.")
        print("\nYou have two options:")
        print("\nOption A - Quick Method:")
        print("1. Go to https://training.grantcardone.com/library")
        print("2. Run this JavaScript in console:")
        print("-" * 30)

        quick_js = """// Quick Video Extraction
(function() {
    const videos = [];
    document.querySelectorAll('video, iframe').forEach((el, i) => {
        const url = el.src || el.querySelector('source')?.src;
        if (url) videos.push({url, title: el.title || `Video ${i+1}`});
    });
    console.log(`Found ${videos.length} videos:`, videos);
    const dataStr = JSON.stringify({videos, timestamp: new Date().toISOString()}, null, 2);
    const blob = new Blob([dataStr], {type: 'application/json'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a'); a.href = url; a.download = 'grantcardone_videos.json'; a.click();
})();"""

        print(quick_js)

        print("\nOption B - Complete Automation:")
        print("1. Navigate through your course pages")
        print("2. Run extraction on each page with videos")
        print("3. Collect all video URLs")

        print("\n📁 Save results as: grantcardone_videos.json")
        input("Press Enter when you have grantcardone_videos.json...")

    def check_stage2_complete(self):
        """Check if stage2 data exists"""
        videos_file = self.base_dir / "grantcardone_videos.json"
        complete_file = self.base_dir / "grantcardone_complete.json"

        # Check either file
        for file_path in [complete_file, videos_file]:
            if file_path.exists():
                try:
                    with open(file_path, 'r') as f:
                        data = json.load(f)
                        videos = data.get('videos', [])
                        if videos:
                            print(f"✅ Stage 2 complete: {len(videos)} videos found")
                            return videos
                except json.JSONDecodeError:
                    pass

        print("❌ Stage 2 data not found or invalid")
        return None

    def create_aria2_input_file(self, videos):
        """Create aria2c input file for downloads"""
        input_file = self.base_dir / "aria2_downloads.txt"

        with open(input_file, 'w') as f:
            for i, video in enumerate(videos, 1):
                title = video.get('title', f'Video {i}')
                safe_title = re.sub(r'[^\w\s-]', '', title).strip()
                safe_title = re.sub(r'[-\s]+', '-', safe_title)
                filename = f"{i:03d}-{safe_title}"

                url = video['url']
                f.write(f"{url}\n")
                f.write(f"  out={filename}.%(ext)s\nn")
                f.write("\n")

        print(f"✅ aria2c input file created: {input_file}")
        return input_file

    def download_videos(self, videos):
        """Download videos using aria2c"""
        if not videos:
            print("❌ No videos to download")
            return

        print(f"\n🚀 DOWNLOAD STAGE: {len(videos)} videos with 16 parallel streams")
        print("=" * 60)

        # Setup aria2c
        self.check_aria2c()
        config_path = self.create_aria2_config()
        input_file = self.create_aria2_input_file(videos)

        # aria2c command
        cmd = [
            'aria2c',
            f'--conf-path={config_path}',
            f'--input-file={input_file}',
            '--show-console-readout=true',
            '--summary-interval=30',
            '--human-readable=true'
        ]

        try:
            print("🔥 Starting aria2c with maximum speed...")
            print("Each file will download with 16 parallel connections")
            print("Press Ctrl+C to pause/stop downloads\n")

            # Run aria2c
            subprocess.run(cmd, cwd=self.base_dir)

            print("\n✅ DOWNLOAD COMPLETE!")
            print(f"📁 All videos saved to: {self.download_dir}")

            # Show summary
            video_files = list(self.download_dir.glob("*.mp4")) + list(self.download_dir.glob("*.m4v")) + list(self.download_dir.glob("*.webm"))
            print(f"📊 Downloaded files: {len(video_files)}")

            if video_files:
                total_size = sum(f.stat().st_size for f in video_files)
                size_gb = total_size / (1024**3)
                print(f"💾 Total size: {size_gb:.2f} GB")

        except KeyboardInterrupt:
            print("\n⏸️ Downloads paused. Run again to resume.")
        except FileNotFoundError:
            print("❌ aria2c command failed. Please check installation.")
        except Exception as e:
            print(f"❌ Download error: {e}")

    def run(self):
        """Main execution"""
        print("🎥 Grant Cardone Video Downloader - Master System")
        print("=" * 60)
        print("🚀 Ultra-fast downloads with aria2c (16 parallel streams)")
        print("📁 Output:", self.download_dir)

        # Check if we already have video data
        videos = self.check_stage2_complete()

        if not videos:
            # Try stage 1 first
            courses = self.check_stage1_complete()
            if not courses:
                self.stage1_instructions()
                courses = self.check_stage1_complete()

            # Then stage 2
            if courses:
                print(f"📚 Found {len(courses)} courses from Stage 1")
                self.stage2_instructions()
                videos = self.check_stage2_complete()

        # Download videos
        if videos:
            self.download_videos(videos)
        else:
            print("\n❌ No video data found. Please complete the extraction steps.")
            print("\n📋 QUICK START:")
            print("1. Go to https://training.grantcardone.com/library")
            print("2. Run the JavaScript from stage1_library_scraper.js")
            print("3. Run the JavaScript from stage2_automated_scraper.py")
            print("4. Run this master script again")

if __name__ == "__main__":
    master = GrantCardoneMaster()
    master.run()