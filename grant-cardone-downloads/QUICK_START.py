#!/usr/bin/env python3
"""
Grant Cardone Video Downloader - Quick Start
Simple instructions without interactive input
"""

import os
import json
import subprocess
import re
from pathlib import Path

class QuickStart:
    def __init__(self):
        self.base_dir = Path(__file__).parent
        self.download_dir = self.base_dir / "grant-cardone-downloads"
        self.download_dir.mkdir(exist_ok=True)

    def show_instructions(self):
        print("🎥 GRANT CARDONE VIDEO DOWNLOADER - QUICK START")
        print("=" * 60)
        print("🚀 16 parallel streams with aria2c for maximum speed!")
        print()

    def check_files(self):
        """Check for existing data files"""
        courses_file = self.base_dir / "grantcardone_courses.json"
        videos_file = self.base_dir / "grantcardone_videos.json"
        complete_file = self.base_dir / "grantcardone_complete.json"

        files_found = []
        if courses_file.exists():
            files_found.append("✅ grantcardone_courses.json")
        if videos_file.exists():
            files_found.append("✅ grantcardone_videos.json")
        if complete_file.exists():
            files_found.append("✅ grantcardone_complete.json")

        if files_found:
            print("📁 Found files:")
            for file in files_found:
                print(f"   {file}")
            print()

        return files_found

    def stage1_js(self):
        """Show Stage 1 JavaScript"""
        print("🎯 STEP 1: Extract Course Links")
        print("-" * 40)
        print("1. Chrome → https://training.grantcardone.com/library")
        print("2. DevTools (Cmd+Opt+I) → Console")
        print("3. Paste this JavaScript:")
        print()
        print("// QUICK START - Extract all courses with 'Start Now' buttons")
        print("(function() {")
        print("    const courses = [];")
        print("    document.querySelectorAll('button, a').forEach(btn => {")
        print("        const text = btn.textContent?.toLowerCase() || '';")
        print("        if (text.includes('start now') || text.includes('start')) {")
        print("            const container = btn.closest('.course, .lesson, .program, [class*=\"course\"]');")
        print("            const title = container?.querySelector('h1, h2, h3, .title')?.textContent?.trim() || btn.textContent.trim();")
        print("            const link = container?.querySelector('a[href]')?.href || btn.href || '';")
        print("            courses.push({id: courses.length + 1, title, link});")
        print("        }")
        print("    });")
        print("    console.log(`Found ${courses.length} courses:`, courses);")
        print("    const data = JSON.stringify({courses, timestamp: new Date().toISOString()}, null, 2);")
        print("    const blob = new Blob([data], {type: 'application/json'});")
        print("    const url = URL.createObjectURL(blob);")
        print("    const a = document.createElement('a'); a.href = url; a.download = 'grantcardone_courses.json'; a.click();")
        print("    console.log('✅ Saved as grantcardone_courses.json');")
        print("})();")
        print()

    def stage2_js(self):
        """Show Stage 2 JavaScript"""
        print("🎯 STEP 2: Extract Video URLs")
        print("-" * 40)
        print("1. Navigate to pages with videos")
        print("2. Run this in console:")
        print()
        print("// QUICK START - Extract video URLs")
        print("(function() {")
        print("    const videos = [];")
        print("    document.querySelectorAll('video').forEach(video => {")
        print("        const src = video.src || video.querySelector('source')?.src;")
        print("        if (src) videos.push({url: src, title: video.title || `Video ${videos.length + 1}`});")
        print("    });")
        print("    document.querySelectorAll('iframe[src*=\"vimeo\"], iframe[src*=\"youtube\"]').forEach(iframe => {")
        print("        videos.push({url: iframe.src, title: iframe.title || `Video ${videos.length + 1}`});")
        print("    });")
        print("    console.log(`Found ${videos.length} videos:`, videos);")
        print("    const data = JSON.stringify({videos, timestamp: new Date().toISOString()}, null, 2);")
        print("    const blob = new Blob([data], {type: 'application/json'});")
        print("    const url = URL.createObjectURL(blob);")
        print("    const a = document.createElement('a'); a.href = url; a.download = 'grantcardone_videos.json'; a.click();")
        print("    console.log('✅ Saved as grantcardone_videos.json');")
        print("})();")
        print()

    def setup_aria2c(self):
        """Setup aria2c configuration"""
        print("🚀 STEP 3: Configure aria2c")
        print("-" * 40)

        config = f"""# Grant Cardone Ultra-Fast Configuration
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
dir={self.download_dir.absolute()}
log-level=notice
show-console-readout=true
disk-cache=64M
enable-http-keep-alive=true
enable-http-pipelining=true
"""

        config_path = self.base_dir / "aria2.conf"
        with open(config_path, 'w') as f:
            f.write(config)

        print(f"✅ Configuration saved: {config_path}")
        print("🔥 16 parallel connections per file!")
        print("📄 1MB chunks for maximum speed")
        print()

    def download_with_aria2c(self):
        """Download videos using aria2c"""
        videos_file = self.base_dir / "grantcardone_videos.json"
        complete_file = self.base_dir / "grantcardone_complete.json"

        # Find videos file
        video_file = None
        if complete_file.exists():
            video_file = complete_file
        elif videos_file.exists():
            video_file = videos_file

        if not video_file:
            print("❌ No video data file found. Complete Steps 1 & 2 first.")
            return

        # Load videos
        try:
            with open(video_file, 'r') as f:
                data = json.load(f)
                videos = data.get('videos', [])
        except (json.JSONDecodeError, FileNotFoundError):
            print("❌ Invalid video data file")
            return

        if not videos:
            print("❌ No videos found in data file")
            return

        print(f"🎥 STEP 4: Download {len(videos)} videos")
        print("-" * 40)

        # Create aria2c input file
        input_file = self.base_dir / "aria2_downloads.txt"
        with open(input_file, 'w') as f:
            for i, video in enumerate(videos, 1):
                title = video.get('title', f'Video {i}')
                safe_title = re.sub(r'[^\w\s-]', '', title).strip()
                safe_title = re.sub(r'[-\s]+', '-', safe_title)
                filename = f"{i:03d}-{safe_title}"

                f.write(f"{video['url']}\n")
                f.write(f"  out={filename}.%(ext)s\n")
                f.write("\n")

        print(f"✅ Created download list: {input_file}")
        print("🔥 Starting aria2c with 16 parallel streams...")
        print("⚡ Each file downloads in 1MB chunks across 16 connections")
        print()

        # Run aria2c
        cmd = [
            'aria2c',
            '--conf-path=aria2.conf',
            '--input-file=aria2_downloads.txt',
            '--show-console-readout=true',
            '--summary-interval=30',
            '--human-readable=true'
        ]

        try:
            subprocess.run(cmd, cwd=self.base_dir)
            print("\n🎉 DOWNLOAD COMPLETE!")
            print(f"📁 Videos saved to: {self.download_dir}")
        except FileNotFoundError:
            print("❌ aria2c not found. Install with: brew install aria2c")
        except Exception as e:
            print(f"❌ Error: {e}")

    def run(self):
        """Main execution"""
        self.show_instructions()

        # Check existing files
        existing_files = self.check_files()

        if "✅ grantcardone_complete.json" in existing_files or "✅ grantcardone_videos.json" in existing_files:
            print("🎥 Video data found! Proceeding to download...")
            self.setup_aria2c()
            self.download_with_aria2c()
        else:
            print("📋 FOLLOW THESE STEPS:")
            print()
            self.stage1_js()
            self.stage2_js()
            self.setup_aria2c()
            print("🔄 After completing Steps 1 & 2, run this script again to download!")

if __name__ == "__main__":
    quick = QuickStart()
    quick.run()