// Grant Cardone Network Video Extractor - COPY THIS ENTIRE CODE
(function() {
    console.log('🌐 Starting Network Video Extraction...');

    const videos = [];
    const networkRequests = [];
    let processedCount = 0;

    // Intercept network requests
    const originalFetch = window.fetch;
    const originalXHROpen = XMLHttpRequest.prototype.open;
    const originalXHRSend = XMLHttpRequest.prototype.send;

    // Track fetch requests
    window.fetch = function(...args) {
        const url = args[0];
        const options = args[1] || {};

        if (typeof url === 'string') {
            networkRequests.push({
                type: 'fetch',
                url: url,
                method: options.method || 'GET',
                timestamp: Date.now()
            });
        }

        return originalFetch.apply(this, args).then(response => {
            if (url.includes('video') || url.includes('stream') || url.includes('vimeo') || url.includes('youtube')) {
                response.clone().text().then(body => {
                    try {
                        const data = JSON.parse(body);
                        if (data.url || data.video_url || data.stream_url) {
                            videos.push({
                                url: data.url || data.video_url || data.stream_url,
                                title: data.title || data.name || `Video ${videos.length + 1}`,
                                source: 'network-fetch'
                            });
                        }
                    } catch (e) {
                        // Not JSON, ignore
                    }
                });
            }
            return response;
        });
    };

    // Track XHR requests
    XMLHttpRequest.prototype.open = function(method, url, ...args) {
        this._url = url;
        this._method = method;

        networkRequests.push({
            type: 'xhr',
            url: url,
            method: method,
            timestamp: Date.now()
        });

        return originalXHROpen.apply(this, [method, url, ...args]);
    };

    XMLHttpRequest.prototype.send = function(data) {
        const xhr = this;

        const originalOnReadyStateChange = xhr.onreadystatechange;
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4 && xhr.status === 200) {
                try {
                    const responseText = xhr.responseText;
                    const url = xhr._url;

                    if (url.includes('video') || url.includes('stream') || url.includes('vimeo') || url.includes('youtube')) {
                        let responseData;
                        try {
                            responseData = JSON.parse(responseText);

                            if (responseData.videos && Array.isArray(responseData.videos)) {
                                responseData.videos.forEach(video => {
                                    videos.push({
                                        url: video.url || video.video_url || video.stream_url,
                                        title: video.title || video.name || `Video ${videos.length + 1}`,
                                        source: 'network-xhr-videos-array'
                                    });
                                });
                            } else if (responseData.url || responseData.video_url) {
                                videos.push({
                                    url: responseData.url || responseData.video_url,
                                    title: responseData.title || responseData.name || `Video ${videos.length + 1}`,
                                    source: 'network-xhr-single'
                                });
                            }
                        } catch (e) {
                            // Not JSON, check if response text contains URLs
                            const urlMatches = responseText.match(/https?:\/\/[^\s"']+\.(?:mp4|m3u8|webm|avi|mov)/gi);
                            if (urlMatches) {
                                urlMatches.forEach(url => {
                                    videos.push({
                                        url: url,
                                        title: `Video ${videos.length + 1}`,
                                        source: 'network-xhr-text'
                                    });
                                });
                            }
                        }
                    }
                } catch (e) {
                    console.log('Error processing XHR response:', e);
                }
            }

            if (originalOnReadyStateChange) {
                originalOnReadyStateChange.apply(this, arguments);
            }
        };

        return originalXHRSend.apply(this, [data]);
    };

    // Extract existing videos from the page
    function extractExistingVideos() {
        console.log('🎥 Scanning page for existing videos...');

        // Look for video elements
        document.querySelectorAll('video').forEach((video, index) => {
            if (video.src || video.querySelector('source')) {
                const src = video.src || video.querySelector('source').src;
                if (src && src.startsWith('http')) {
                    videos.push({
                        url: src,
                        title: video.title || video.getAttribute('data-title') || `Video ${videos.length + 1}`,
                        source: 'video-element'
                    });
                }
            }
        });

        // Look for iframes with video sources
        document.querySelectorAll('iframe').forEach((iframe, index) => {
            const src = iframe.src;
            if (src && (src.includes('vimeo.com') || src.includes('youtube.com') || src.includes('player'))) {
                videos.push({
                    url: src,
                    title: iframe.title || iframe.getAttribute('data-title') || `Video ${videos.length + 1}`,
                    source: 'iframe'
                });
            }
        });

        // Look for script tags with video data
        document.querySelectorAll('script').forEach((script, index) => {
            const text = script.textContent;

            // Look for JSON data containing video URLs
            const jsonPatterns = [
                /{[\s\S]*?"video(?:_url)?"[\s\S]*?}/g,
                /{[\s\S]*?"stream(?:_url)?"[\s\S]*?}/g,
                /{[\s\S]*?"url"[\s\S]*?"mp4"[\s\S]*?}/g
            ];

            jsonPatterns.forEach(pattern => {
                const matches = text.match(pattern);
                if (matches) {
                    matches.forEach(match => {
                        try {
                            const data = JSON.parse(match);
                            const videoUrl = data.video_url || data.stream_url || data.url;
                            const title = data.title || data.name || data.lesson_title;

                            if (videoUrl && videoUrl.startsWith('http')) {
                                videos.push({
                                    url: videoUrl,
                                    title: title || `Video ${videos.length + 1}`,
                                    source: 'script-json'
                                });
                            }
                        } catch (e) {
                            // Invalid JSON, ignore
                        }
                    });
                }
            });

            // Look for direct URLs in scripts
            const urlPatterns = [
                /https?:\/\/[^\s"']+\.(?:mp4|m3u8|webm|avi|mov)(?=["\s'])/g,
                /["']((?:https?:\/\/)?(?:vimeo\.com|youtube\.com)\/[^"']+)["']/g
            ];

            urlPatterns.forEach(pattern => {
                const matches = text.match(pattern);
                if (matches) {
                    matches.forEach(url => {
                        url = url.replace(/["']/g, '');
                        if (url.startsWith('http') && (url.includes('mp4') || url.includes('m3u8') || url.includes('vimeo') || url.includes('youtube'))) {
                            videos.push({
                                url: url,
                                title: `Video ${videos.length + 1}`,
                                source: 'script-url'
                            });
                        }
                    });
                }
            });
        });

        // Look for data attributes and JavaScript variables
        const possibleSelectors = [
            '[data-video-url]',
            '[data-stream-url]',
            '[data-lesson-video]',
            '.video-url',
            '.stream-url',
            '[onclick*="video"]',
            '[onclick*="play"]'
        ];

        possibleSelectors.forEach(selector => {
            document.querySelectorAll(selector).forEach(element => {
                const videoUrl = (
                    element.getAttribute('data-video-url') ||
                    element.getAttribute('data-stream-url') ||
                    element.getAttribute('data-lesson-video')
                );

                if (videoUrl && videoUrl.startsWith('http')) {
                    videos.push({
                        url: videoUrl,
                        title: element.textContent.trim() || `Video ${videos.length + 1}`,
                        source: 'data-attribute'
                    });
                }
            });
        });

        // Look for course/lesson structure
        const courseElements = document.querySelectorAll([
            '.course-item',
            '.lesson-item',
            '.video-item',
            '[class*="course"]',
            '[class*="lesson"]',
            '[class*="video"]'
        ].join(','));

        courseElements.forEach(element => {
            const link = element.querySelector('a');
            if (link && link.href) {
                const title = (
                    element.querySelector('.title')?.textContent ||
                    element.querySelector('.lesson-title')?.textContent ||
                    element.querySelector('.video-title')?.textContent ||
                    link.textContent.trim()
                );

                if (link.href && (
                    link.href.includes('lesson') ||
                    link.href.includes('video') ||
                    link.href.includes('course')
                )) {
                    videos.push({
                        url: link.href,
                        title: title || `Video ${videos.length + 1}`,
                        source: 'course-structure'
                    });
                }
            }
        });
    }

    // Initial extraction
    extractExistingVideos();

    // Wait a moment for any async loading, then extract more
    setTimeout(() => {
        extractExistingVideos();

        console.log('\n' + '='.repeat(60));
        console.log('🎥 NETWORK EXTRACTION RESULTS');
        console.log('='.repeat(60));
        console.log('📊 Total videos found: ' + videos.length);
        console.log('🌐 Network requests captured: ' + networkRequests.length);

        // Remove duplicates
        const uniqueVideos = videos.filter((video, index, self) =>
            index === self.findIndex((v) => v.url === video.url)
        );

        console.log('🎬 Unique videos: ' + uniqueVideos.length);

        console.log('\n📋 VIDEO LIST:');
        uniqueVideos.forEach((video, index) => {
            console.log((index + 1) + '. ' + video.title);
            console.log('   URL: ' + video.url);
            console.log('   Source: ' + video.source);
            console.log('');
        });

        // Generate download commands
        console.log('\n📥 DOWNLOAD COMMANDS:');
        console.log('='.repeat(30));
        uniqueVideos.forEach((video, index) => {
            const safeTitle = video.title.replace(/[^a-zA-Z0-9\s-]/g, '').trim().replace(/\s+/g, '-');
            const filename = String(index + 1).padStart(3, '0') + '-' + safeTitle;
            console.log('yt-dlp --output "' + filename + '.%(ext)s" --format "best[height<=1080]" "' + video.url + '"');
        });

        // Create downloadable JSON
        const downloadData = {
            videos: uniqueVideos,
            networkRequests: networkRequests,
            timestamp: new Date().toISOString()
        };

        console.log('\n📋 COPY THIS JSON:');
        console.log('='.repeat(30));
        console.log(JSON.stringify(downloadData, null, 2));

        // Also download the file
        const dataStr = JSON.stringify(downloadData, null, 2);
        const dataBlob = new Blob([dataStr], {type: 'application/json'});
        const url = URL.createObjectURL(dataBlob);
        const link = document.createElement('a');
        link.href = url;
        link.download = 'grantcardone_videos.json';
        link.click();
        URL.revokeObjectURL(url);

        console.log('\n✅ Video data downloaded to grantcardone_videos.json');

        return uniqueVideos;
    }, 3000);

    // Return the current video count
    return {
        initialVideos: videos.length,
        networkRequests: networkRequests.length
    };
})();