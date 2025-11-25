// FINAL GRANT CARDONE VIDEO EXTRACTOR - Auto-saves JSON
(function() {
    console.log('🎥 Starting Grant Cardone Video Extraction...');

    const videos = [];

    // Look for video elements
    document.querySelectorAll('video').forEach((video, i) => {
        if (video.src) {
            videos.push({url: video.src, title: video.title || `Video ${i+1}`, source: 'video-element'});
        }
        const source = video.querySelector('source');
        if (source && source.src) {
            videos.push({url: source.src, title: video.title || `Video ${i+1}`, source: 'video-source'});
        }
    });

    // Look for iframes (Vimeo, YouTube, etc.)
    document.querySelectorAll('iframe').forEach((iframe, i) => {
        if (iframe.src && (iframe.src.includes('vimeo') || iframe.src.includes('youtube') || iframe.src.includes('player'))) {
            videos.push({url: iframe.src, title: iframe.title || `Video ${i+1}`, source: 'iframe'});
        }
    });

    // Look for lesson/course links
    document.querySelectorAll('a[href*="lesson"], a[href*="video"], .lesson-item a, .video-item a, .course-item a').forEach((link, i) => {
        const title = link.textContent.trim() || link.title || `Lesson ${i+1}`;
        videos.push({url: link.href, title: title, type: 'lesson-link', source: 'course-structure'});
    });

    // Look for data attributes with video URLs
    document.querySelectorAll('[data-video-url], [data-stream-url], [data-lesson-video]').forEach((element, i) => {
        const videoUrl = element.getAttribute('data-video-url') || element.getAttribute('data-stream-url') || element.getAttribute('data-lesson-video');
        if (videoUrl) {
            const title = element.textContent.trim() || element.title || `Video ${i+1}`;
            videos.push({url: videoUrl, title: title, source: 'data-attribute'});
        }
    });

    // Look for script tags with video data
    document.querySelectorAll('script').forEach((script, i) => {
        const text = script.textContent;

        // Look for video URLs in scripts
        const urlPatterns = [
            /https?:\/\/[^\s"']+\.(?:mp4|m3u8|webm|avi|mov)(?=["\s])/g,
            /https?:\/\/(?:vimeo\.com|youtube\.com|youtu\.be)\/[^\s"']+/g
        ];

        urlPatterns.forEach(pattern => {
            const matches = text.match(pattern);
            if (matches) {
                matches.forEach(url => {
                    videos.push({url: url.trim(), title: `Video ${videos.length + 1}`, source: 'script-detection'});
                });
            }
        });

        // Look for JSON data with video info
        const jsonMatches = text.match(/\{[^{}]*"video[^"]*"[^{}]*\}|\{[^{}]*"url"[^{}]*"video"[^{}]*\}/g);
        if (jsonMatches) {
            jsonMatches.forEach(match => {
                try {
                    const data = JSON.parse(match);
                    const videoUrl = data.video_url || data.stream_url || data.url;
                    const title = data.title || data.name || `Video ${videos.length + 1}`;
                    if (videoUrl && videoUrl.startsWith('http')) {
                        videos.push({url: videoUrl, title: title, source: 'script-json'});
                    }
                } catch (e) {
                    // Invalid JSON, ignore
                }
            });
        }
    });

    // Remove duplicates
    const uniqueVideos = videos.filter((video, index, self) =>
        index === self.findIndex((v) => v.url === video.url)
    );

    console.log('\n' + '='.repeat(60));
    console.log('🎥 GRANT CARDONE VIDEO EXTRACTION RESULTS');
    console.log('='.repeat(60));
    console.log(`📊 Total videos found: ${uniqueVideos.length}`);

    console.log('\n📋 VIDEO LIST:');
    uniqueVideos.forEach((video, index) => {
        console.log(`${index + 1}. ${video.title}`);
        console.log(`   URL: ${video.url}`);
        console.log(`   Source: ${video.source}`);
        console.log('');
    });

    // Create downloadable JSON
    const downloadData = {
        videos: uniqueVideos,
        timestamp: new Date().toISOString(),
        totalVideos: uniqueVideos.length,
        pageUrl: window.location.href
    };

    // Auto-download the JSON file
    const dataStr = JSON.stringify(downloadData, null, 2);
    const dataBlob = new Blob([dataStr], {type: 'application/json'});
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'grantcardone_videos.json';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    console.log('✅ JSON file auto-downloaded as grantcardone_videos.json');
    console.log('📁 Save this file to your download directory');
    console.log('🚀 Then run: python3 interactive_extractor.py');

    // Also show the data in console for manual copy
    console.log('\n📋 COPY THIS JSON:');
    console.log('='.repeat(30));
    console.log(dataStr);

    return {
        videos: uniqueVideos,
        totalVideos: uniqueVideos.length,
        autoDownloaded: true
    };
})();