#!/usr/bin/env python3
"""
Advanced Grant Cardone Video Downloader
Uses browser auth tokens and mimics exact browser behavior
"""

import os
import sys
import json
import time
import random
import requests
import base64
import subprocess
import re
from urllib.parse import urljoin, urlparse, parse_qs
import sqlite3
import tempfile
import shutil
from pathlib import Path

class AdvancedCardoneExtractor:
    def __init__(self):
        self.session = requests.Session()
        self.base_url = "https://training.grantcardone.com"
        self.api_base = "https://training.grantcardone.com/api"
        self.download_dir = "./grant-cardone-downloads"
        self.chrome_profile_dir = None

        # Set up realistic headers to mimic browser
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
            'Accept-Language': 'en-US,en;q=0.9',
            'Accept-Encoding': 'gzip, deflate, br',
            'DNT': '1',
            'Connection': 'keep-alive',
            'Sec-Fetch-Dest': 'document',
            'Sec-Fetch-Mode': 'navigate',
            'Sec-Fetch-Site': 'none',
            'Sec-Fetch-User': '?1',
            'Cache-Control': 'max-age=0',
            'sec-ch-ua': '"Chromium";v="142", "Google Chrome";v="142", "Not(A:Brand";v="24"',
            'sec-ch-ua-mobile': '?0',
            'sec-ch-ua-platform': '"macOS"',
        })

    def find_chrome_profile(self):
        """Find the Chrome profile directory"""
        home = Path.home()
        possible_profiles = [
            home / "Library/Application Support/Google/Chrome",
            home / "Library/Application Support/Google/Chrome/Default",
            home / "Library/Application Support/Google/Chrome/Profile 1",
            home / "Library/Application Support/Google/Chrome/Profile 2",
        ]

        for profile_dir in possible_profiles:
            if (profile_dir / "Cookies").exists():
                self.chrome_profile_dir = profile_dir
                print(f"Found Chrome profile: {profile_dir}")
                return profile_dir

        print("Chrome profile not found")
        return None

    def extract_cookies_from_chrome(self):
        """Extract cookies from Chrome's SQLite database"""
        if not self.chrome_profile_dir:
            self.find_chrome_profile()

        if not self.chrome_profile_dir:
            return False

        cookies_db = self.chrome_profile_dir / "Cookies"

        # Create temporary copy of cookies database
        temp_dir = tempfile.mkdtemp()
        temp_cookies = os.path.join(temp_dir, "Cookies")

        try:
            shutil.copy2(cookies_db, temp_cookies)

            # Connect to copied database
            conn = sqlite3.connect(temp_cookies)
            conn.row_factory = sqlite3.Row

            # Extract relevant cookies for grantcardone.com
            cursor = conn.execute("""
                SELECT name, value, host_key, path, expires_utc, is_secure, is_httponly
                FROM cookies
                WHERE host_key LIKE '%grantcardone%' OR host_key LIKE '%training%'
            """)

            cookies_found = False
            for row in cursor.fetchall():
                cookie_name = row['name']
                cookie_value = row['value']
                domain = row['host_key'].lstrip('.')

                if cookie_value:  # Only add non-empty cookies
                    self.session.cookies.set(cookie_name, cookie_value, domain=domain)
                    print(f"Extracted cookie: {cookie_name} for domain {domain}")
                    cookies_found = True

            conn.close()
            shutil.rmtree(temp_dir)

            return cookies_found

        except Exception as e:
            print(f"Error extracting cookies: {e}")
            if os.path.exists(temp_dir):
                shutil.rmtree(temp_dir)
            return False

    def extract_local_storage(self):
        """Extract localStorage data from Chrome"""
        if not self.chrome_profile_dir:
            return {}

        local_storage_path = self.chrome_profile_dir / "Local Storage/leveldb"

        try:
            # This is simplified - in reality, Chrome's localStorage is in LevelDB format
            # For a complete solution, we'd need to parse the LevelDB files
            # For now, we'll use Chrome's remote debugging protocol

            return {}
        except Exception as e:
            print(f"Error extracting localStorage: {e}")
            return {}

    def get_auth_tokens_via_debugging(self):
        """Get auth tokens using Chrome's remote debugging"""
        try:
            # Check if Chrome is running with remote debugging
            response = requests.get("http://localhost:9222/json", timeout=5)
            if response.status_code != 200:
                return self.start_debugging_chrome()

            tabs = response.json()
            for tab in tabs:
                if "training.grantcardone.com" in tab.get("url", ""):
                    return self.extract_tokens_from_tab(tab)

            return {}

        except Exception as e:
            print(f"Chrome debugging not available: {e}")
            return {}

    def start_debugging_chrome(self):
        """Start Chrome with remote debugging enabled"""
        try:
            print("Starting Chrome with remote debugging...")

            # Kill existing Chrome processes
            subprocess.run(["pkill", "Google Chrome"], capture_output=True)
            time.sleep(2)

            # Start Chrome with debugging
            chrome_cmd = [
                "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
                "--remote-debugging-port=9222",
                "--no-first-run",
                "--no-default-browser-check",
                f"--user-data-dir={self.chrome_profile_dir}" if self.chrome_profile_dir else "",
                "https://training.grantcardone.com/library"
            ]

            # Filter out empty strings
            chrome_cmd = [arg for arg in chrome_cmd if arg]

            subprocess.Popen(chrome_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

            print("Waiting for Chrome to start...")
            time.sleep(5)

            return self.get_auth_tokens_via_debugging()

        except Exception as e:
            print(f"Error starting debugging Chrome: {e}")
            return {}

    def extract_tokens_from_tab(self, tab):
        """Extract auth tokens from a specific Chrome tab"""
        websocket_url = tab.get("webSocketDebuggerUrl")
        if not websocket_url:
            print("No WebSocket URL found")
            return {}

        # For now, use JavaScript execution via debugger protocol
        try:
            js_code = """
            (function() {
                const tokens = {};

                // Get localStorage data
                if (typeof localStorage !== 'undefined') {
                    for (let i = 0; i < localStorage.length; i++) {
                        const key = localStorage.key(i);
                        if (key.includes('token') || key.includes('auth') || key.includes('session')) {
                            tokens[key] = localStorage.getItem(key);
                        }
                    }
                }

                // Get sessionStorage data
                if (typeof sessionStorage !== 'undefined') {
                    for (let i = 0; i < sessionStorage.length; i++) {
                        const key = sessionStorage.key(i);
                        if (key.includes('token') || key.includes('auth') || key.includes('session')) {
                            tokens['session_' + key] = sessionStorage.getItem(key);
                        }
                    }
                }

                // Get cookies that might contain auth tokens
                const cookies = document.cookie.split(';').map(c => c.trim()).reduce((acc, cookie) => {
                    const [name, value] = cookie.split('=');
                    if (name.includes('token') || name.includes('auth') || name.includes('session')) {
                        acc[name] = value;
                    }
                    return acc;
                }, {});

                // Look for auth data in window object
                if (window.user || window.auth || window.token) {
                    tokens.window_data = {
                        user: window.user,
                        auth: window.auth,
                        token: window.token
                    };
                }

                // Look for common JavaScript variables
                const commonVars = ['jwt', 'accessToken', 'authToken', 'apiToken'];
                commonVars.forEach(varName => {
                    if (window[varName]) {
                        tokens[varName] = window[varName];
                    }
                });

                return tokens;
            })();
            """

            # Use Chrome DevTools Protocol to execute JavaScript
            # This would require a WebSocket client implementation
            # For simplicity, we'll ask the user to run this manually

            print("Manual token extraction required:")
            print("1. Open Chrome DevTools (Cmd+Opt+I)")
            print("2. Go to the Grant Cardone training tab")
            print("3. Go to Console tab")
            print("4. Run this JavaScript:")
            print(js_code)
            print("5. Copy the output and save to 'auth_tokens.json'")

            return {}

        except Exception as e:
            print(f"Error extracting tokens: {e}")
            return {}

    def load_auth_tokens(self):
        """Load auth tokens from file"""
        try:
            with open('grantcardone_auth.json', 'r') as f:
                tokens = json.load(f)

            # Apply tokens to session headers
            if 'headers' in tokens:
                for key, value in tokens['headers'].items():
                    if value:  # Only add non-empty values
                        self.session.headers[key] = value

            # Add all cookies from auth data
            if 'cookies' in tokens:
                for key, value in tokens['cookies'].items():
                    if value:  # Only add non-empty cookies
                        self.session.cookies.set(key, value, domain='.grantcardone.com')
                        self.session.cookies.set(key, value, domain='.training.grantcardone.com')

            # Add localStorage items as headers if they look like auth tokens
            if 'localStorage' in tokens:
                for key, value in tokens['localStorage'].items():
                    if 'token' in key.lower() and value:
                        if not value.startswith('Bearer '):
                            self.session.headers['Authorization'] = f"Bearer {value}"
                        else:
                            self.session.headers['Authorization'] = value

            print(f"Loaded {len(tokens)} auth tokens")
            return True

        except FileNotFoundError:
            print("grantcardone_auth.json not found - please extract tokens first")
            return False
        except json.JSONDecodeError:
            print("Invalid JSON in grantcardone_auth.json")
            return False

    def random_delay(self, min_seconds=1, max_seconds=3):
        """Add random delay to mimic human behavior"""
        delay = random.uniform(min_seconds, max_seconds)
        time.sleep(delay)

    def make_authenticated_request(self, url, method='GET', data=None, params=None):
        """Make an authenticated request with proper delays"""
        self.random_delay()

        try:
            if method == 'GET':
                response = self.session.get(url, params=params, timeout=30)
            elif method == 'POST':
                response = self.session.post(url, json=data, params=params, timeout=30)
            else:
                raise ValueError(f"Unsupported method: {method}")

            return response

        except Exception as e:
            print(f"Request error: {e}")
            return None

    def discover_api_endpoints(self):
        """Discover API endpoints by analyzing the page"""
        response = self.make_authenticated_request(f"{self.base_url}/library")

        if not response or response.status_code != 200:
            print("Failed to access library page")
            return []

        # Parse HTML to find API calls
        html_content = response.text

        # Look for common API patterns
        api_patterns = [
            r'/api/[^"\']+',
            r'https://[^"\']*training\.grantcardone\.com/api/[^"\']+',
            r'"url":\s*"([^"]*api[^"]*)"',
            r'fetch\(["\']([^"\']+)["\']',
            r'\.get\(["\']([^"\']+)["\']',
            r'\.post\(["\']([^"\']+)["\']',
        ]

        endpoints = set()
        for pattern in api_patterns:
            matches = re.findall(pattern, html_content, re.IGNORECASE)
            for match in matches:
                if 'api' in match.lower() and 'grantcardone' in match.lower():
                    endpoints.add(match)

        return list(endpoints)

    def get_video_library_data(self):
        """Get video library data via API"""
        # Common API endpoints for video platforms
        potential_endpoints = [
            "/api/v1/videos",
            "/api/videos",
            "/api/library",
            "/api/courses",
            "/api/programs",
            "/api/lessons",
            "/api/content",
            "/api/user/library",
            "/api/v2/courses",
            "/api/v2/programs"
        ]

        video_data = []

        for endpoint in potential_endpoints:
            url = urljoin(self.api_base, endpoint)
            print(f"Trying endpoint: {url}")

            response = self.make_authenticated_request(url)

            if response and response.status_code == 200:
                try:
                    data = response.json()

                    # Check if this endpoint returns video data
                    if self.contains_video_data(data):
                        print(f"✅ Found video data at: {endpoint}")
                        videos = self.extract_videos_from_response(data)
                        video_data.extend(videos)

                except json.JSONDecodeError:
                    continue

            # Add small delay between requests
            time.sleep(random.uniform(1, 2))

        return video_data

    def contains_video_data(self, data):
        """Check if response contains video data"""
        video_indicators = ['videos', 'courses', 'lessons', 'programs', 'content']

        if isinstance(data, dict):
            return any(key in data for key in video_indicators)
        elif isinstance(data, list):
            return len(data) > 0
        return False

    def extract_videos_from_response(self, data):
        """Extract video information from API response"""
        videos = []

        def process_item(item, program_name=""):
            if isinstance(item, dict):
                # Look for video URLs and titles
                video_url = (
                    item.get('video_url') or
                    item.get('url') or
                    item.get('stream_url') or
                    item.get('download_url') or
                    item.get('source_url')
                )

                title = (
                    item.get('title') or
                    item.get('name') or
                    item.get('lesson_title') or
                    item.get('video_title') or
                    f"Video {len(videos) + 1}"
                )

                if video_url:
                    safe_filename = re.sub(r'[^\w\s-]', '', title).strip()
                    safe_filename = re.sub(r'[-\s]+', '-', safe_filename)

                    videos.append({
                        'title': title,
                        'url': video_url,
                        'filename': f"{len(videos) + 1:03d}-{safe_filename}",
                        'program': program_name,
                        'duration': item.get('duration'),
                        'description': item.get('description')
                    })

                # Recursively process nested items
                for key, value in item.items():
                    if isinstance(value, (dict, list)):
                        process_item(value, program_name)

            elif isinstance(item, list):
                for list_item in item:
                    process_item(list_item, program_name)

        # Start processing
        process_item(data)

        return videos

    def download_videos(self, videos):
        """Download all videos using yt-dlp"""
        if not videos:
            print("No videos to download")
            return

        os.makedirs(self.download_dir, exist_ok=True)

        print(f"Downloading {len(videos)} videos...")

        for i, video in enumerate(videos, 1):
            print(f"\n[{i}/{len(videos)}] {video['title']}")

            # Create program directory if needed
            if video.get('program'):
                program_dir = os.path.join(self.download_dir, video['program'])
                os.makedirs(program_dir, exist_ok=True)
                output_path = os.path.join(program_dir, f"{video['filename']}.%(ext)s")
            else:
                output_path = os.path.join(self.download_dir, f"{video['filename']}.%(ext)s")

            # yt-dlp command with realistic options
            cmd = [
                'yt-dlp',
                '--user-agent', self.session.headers['User-Agent'],
                '--cookies', '/dev/null',  # Don't use cookies, rely on auth headers
                '--add-header', f'Authorization: {self.session.headers.get("Authorization", "")}',
                '--no-warnings',
                '--embed-metadata',
                '--embed-thumbnail',
                '--write-subtitles',
                '--write-auto-subs',
                '--sub-langs', 'en',
                '--output', output_path,
                '--format', 'best[height<=1080]',
                '--retries', '3',
                '--fragment-retries', '3',
                video['url']
            ]

            # Add description if available
            if video.get('description'):
                cmd.extend(['--add-metadata', '--metadata', f'description={video["description"]}'])

            try:
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
                if result.returncode == 0:
                    print(f"✅ Downloaded: {video['title']}")
                else:
                    print(f"❌ Failed: {video['title']}")
                    print(f"Error: {result.stderr}")

            except subprocess.TimeoutExpired:
                print(f"⏰ Timeout: {video['title']}")
            except Exception as e:
                print(f"❌ Error: {video['title']} - {e}")

            # Random delay between downloads to be respectful
            self.random_delay(2, 5)

    def run(self):
        """Main execution method"""
        print("🎥 Advanced Grant Cardone Video Extractor")
        print("=" * 50)

        # Step 1: Extract auth tokens
        print("\n1. Extracting authentication tokens...")

        # Try cookies first
        if self.extract_cookies_from_chrome():
            print("✅ Cookies extracted successfully")
        else:
            print("⚠️  Cookie extraction failed, trying debugging...")
            self.get_auth_tokens_via_debugging()

        # Load manual tokens if available
        self.load_auth_tokens()

        # Step 2: Get video data
        print("\n2. Discovering video data...")
        videos = self.get_video_library_data()

        if not videos:
            print("❌ No video data found")
            print("Please ensure you're logged in and try manual token extraction")
            return

        print(f"✅ Found {len(videos)} videos")

        # Step 3: Download videos
        print("\n3. Starting downloads...")
        self.download_videos(videos)

        print(f"\n🎉 Download complete! Videos saved to: {os.path.abspath(self.download_dir)}")

if __name__ == "__main__":
    extractor = AdvancedCardoneExtractor()
    extractor.run()