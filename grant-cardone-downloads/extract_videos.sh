#!/bin/bash

# Grant Cardone Video Extraction Script
# This script helps extract video URLs from the training platform

echo "🎥 Grant Cardone Video Extractor"
echo "=================================="

# Create download directory
mkdir -p grant-cardone-videos
cd grant-cardone-videos

echo "✅ Created/entered download directory"

# Start Chrome with remote debugging if not already running
echo "🔍 Checking Chrome remote debugging..."

if ! curl -s http://localhost:9222/json > /dev/null 2>&1; then
    echo "⚠️  Chrome remote debugging not detected"
    echo "🚀 Starting Chrome with remote debugging..."

    # Close existing Chrome instances
    osascript -e 'tell application "Google Chrome" to quit'
    sleep 2

    # Start Chrome with remote debugging
    open -a "Google Chrome" --args --remote-debugging-port=9222 --user-data-dir="/tmp/chrome-debug"

    echo "⏳ Waiting for Chrome to start..."
    sleep 3

    # Open Grant Cardone training
    open -a "Google Chrome" "https://training.grantcardone.com/library"

    echo "⏳ Waiting for page to load..."
    sleep 5
else
    echo "✅ Chrome remote debugging is running"
fi

echo ""
echo "📋 INSTRUCTIONS:"
echo "================"
echo "1. In Chrome, open the Grant Cardone training tab"
echo "2. Open Developer Tools (Cmd+Opt+I)"
echo "3. Go to the Console tab"
echo "4. Copy and paste this JavaScript code:"
echo ""

cat << 'EOF'
// Grant Cardone Video Extraction Script
(function() {
    console.log('🎥 Starting video extraction...');

    const videos = [];
    let processedCount = 0;

    // Helper function to sanitize text
    function sanitizeText(text) {
        return text.replace(/[\r\n\t]+/g, ' ').trim();
    }

    // Helper function to create safe filename
    function createSafeFilename(title, index) {
        const safe = title.replace(/[^\w\s-]/g, '').replace(/[-\s]+/g, '-').trim();
        return `${String(index + 1).padStart(3, '0')}-${safe}`.substring(0, 100);
    }

    // Find program/course structure
    console.log('📚 Looking for program structure...');
    const programSections = document.querySelectorAll([
        '.program',
        '.course',
        '.module',
        '.section',
        '[class*="program"]',
        '[class*="course"]',
        '[class*="module"]',
        '[class*="section"]'
    ].join(','));

    const videoData = {
        programs: [],
        totalVideos: 0,
        extractionTime: new Date().toISOString()
    };

    programSections.forEach((section, sectionIndex) => {
        const sectionTitle = section.querySelector('h1, h2, h3, h4, .title, .section-title, .program-title');
        const programTitle = sectionTitle ? sanitizeText(sectionTitle.textContent) : `Program ${sectionIndex + 1}`;

        const videosInSection = [];

        // Find videos within this section
        const videoElements = section.querySelectorAll([
            'a[href*="video"]',
            'a[href*="lesson"]',
            'a[href*="vimeo"]',
            'a[href*="youtube"]',
            '.video-item a',
            '.lesson-item a',
            '.video a',
            '[class*="video"] a',
            '[class*="lesson"] a'
        ].join(','));

        videoElements.forEach((link, videoIndex) => {
            const title = sanitizeText(link.textContent || link.title || `Video ${videoIndex + 1}`);
            const url = link.href;

            if (url && !url.includes('#') && url !== 'javascript:void(0)') {
                videosInSection.push({
                    title: title,
                    url: url,
                    index: videoIndex + 1,
                    filename: createSafeFilename(title, videoData.totalVideos + videoIndex)
                });
            }
        });

        if (videosInSection.length > 0) {
            videoData.programs.push({
                title: programTitle,
                videos: videosInSection,
                videoCount: videosInSection.length
            });
            videoData.totalVideos += videosInSection.length;
        }
    });

    // Also look for individual videos not in programs
    console.log('🔍 Looking for individual videos...');
    const allVideoLinks = document.querySelectorAll('a[href]');
    const individualVideos = [];

    allVideoLinks.forEach((link, index) => {
        const url = link.href;
        const title = sanitizeText(link.textContent || link.title || `Video ${index + 1}`);

        // Check if this looks like a video link
        if (url && (
            url.includes('video') ||
            url.includes('vimeo') ||
            url.includes('youtube') ||
            url.includes('watch') ||
            url.includes('lesson') ||
            url.includes('play')
        ) && !url.includes('#') && url !== 'javascript:void(0)') {

            // Check if already captured in program structure
            const alreadyCaptured = videoData.programs.some(program =>
                program.videos.some(video => video.url === url)
            );

            if (!alreadyCaptured) {
                individualVideos.push({
                    title: title,
                    url: url,
                    index: individualVideos.length + 1,
                    filename: createSafeFilename(title, videoData.totalVideos + individualVideos.length)
                });
            }
        }
    });

    if (individualVideos.length > 0) {
        videoData.programs.push({
            title: 'Individual Videos',
            videos: individualVideos,
            videoCount: individualVideos.length
        });
        videoData.totalVideos += individualVideos.length;
    }

    // Generate download commands
    console.log('📥 Generating download commands...');
    const downloadCommands = [];

    videoData.programs.forEach((program, programIndex) => {
        console.log(`\n📚 Program: ${program.title} (${program.videoCount} videos)`);

        program.videos.forEach((video, videoIndex) => {
            const outputPath = `"${program.title}/${video.filename}.%(ext)s"`;
            const cmd = `yt-dlp --no-warnings --embed-metadata --write-subtitles --sub-langs en --output ${outputPath} --format "best[height<=1080]" "${video.url}"`;
            downloadCommands.push(cmd);
            console.log(`  ${video.index}. ${video.title}`);
            console.log(`     URL: ${video.url}`);
        });
    });

    // Final results
    console.log('\n' + '='.repeat(80));
    console.log('🎥 GRANT CARDONE VIDEO EXTRACTION RESULTS');
    console.log('='.repeat(80));
    console.log(`📚 Total Programs: ${videoData.programs.length}`);
    console.log(`🎥 Total Videos: ${videoData.totalVideos}`);
    console.log(`⏰ Extracted: ${videoData.extractionTime}`);

    // Create downloadable JSON
    const jsonData = {
        ...videoData,
        downloadCommands: downloadCommands
    };

    console.log('\n📋 COPY THIS JSON DATA:');
    console.log('='.repeat(50));
    console.log(JSON.stringify(jsonData, null, 2));
    console.log('='.repeat(50));

    // Also create the download script
    console.log('\n📥 DOWNLOAD SCRIPT:');
    console.log('='.repeat(50));
    console.log('#!/bin/bash');
    console.log('# Grant Cardone Video Download Script');
    console.log('mkdir -p "grant-cardone-downloads"');
    console.log('cd "grant-cardone-downloads"');
    downloadCommands.forEach(cmd => {
        console.log(cmd);
    });
    console.log('echo "✅ Download complete!"');
    console.log('='.repeat(50));

    console.log('\n✅ Extraction complete!');
    console.log(`📊 Found ${videoData.totalVideos} videos across ${videoData.programs.length} programs`);

    return jsonData;
})();
EOF

echo ""
echo "5. Run the JavaScript in the console"
echo "6. Copy the JSON output from the console"
echo "7. Save it to 'video_data.json' in this directory"
echo ""
echo "🚀 Ready to start extraction!"
echo ""