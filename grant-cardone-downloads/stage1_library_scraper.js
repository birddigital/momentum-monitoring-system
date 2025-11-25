// STAGE 1: Grant Cardone Library Scraper - Extract all "Start Now" buttons
(function() {
    console.log('📚 STAGE 1: Scraping Grant Cardone Library...');

    const courses = [];

    // Find all "Start Now" buttons and their associated course info
    const startNowButtons = Array.from(document.querySelectorAll('button, a')).filter(element => {
        const text = (element.textContent || '').trim().toLowerCase();
        return text.includes('start now') || text.includes('start') || text.includes('begin') || text.includes('continue');
    });

    console.log(`Found ${startNowButtons.length} "Start Now" buttons`);

    startNowButtons.forEach((button, index) => {
        // Find the containing course/lesson element
        let courseContainer = button.closest('.course, .lesson, .program, .module, [class*="course"], [class*="lesson"], [class*="program"]');

        if (!courseContainer) {
            // Try other common container patterns
            courseContainer = button.closest('div[class*="item"], div[class*="card"], div[class*="row"], li, article, section');
        }

        // Extract course information
        let title = '';
        let description = '';
        let link = '';

        // Get link if it's an <a> tag or find nearby link
        if (button.tagName === 'A' && button.href) {
            link = button.href;
        } else {
            // Look for links in the container
            const containerLink = courseContainer?.querySelector('a[href]');
            if (containerLink) {
                link = containerLink.href;
            }
        }

        // Extract title from container
        if (courseContainer) {
            const titleElement = courseContainer.querySelector('h1, h2, h3, h4, .title, .course-title, .lesson-title, [class*="title"]');
            if (titleElement) {
                title = titleElement.textContent.trim();
            }

            const descElement = courseContainer.querySelector('.description, .summary, p, [class*="description"], [class*="summary"]');
            if (descElement) {
                description = descElement.textContent.trim().substring(0, 200); // Limit description length
            }
        }

        // If no title found, use button text or generate one
        if (!title) {
            title = button.textContent.trim() || `Course ${index + 1}`;
        }

        // Fallback: try to extract from nearby elements
        if (!title && courseContainer) {
            const textElements = courseContainer.querySelectorAll('h1, h2, h3, h4, h5, h6, .title, strong, b');
            for (let elem of textElements) {
                const text = elem.textContent.trim();
                if (text.length > 3 && text.length < 100) {
                    title = text;
                    break;
                }
            }
        }

        // Clean up the title
        title = title.replace(/^Start Now\s*/i, '').replace(/\s*Start Now$/i, '').trim();
        if (!title) title = `Course ${index + 1}`;

        const course = {
            id: index + 1,
            title: title,
            description: description,
            link: link,
            buttonText: button.textContent.trim(),
            containerClass: courseContainer?.className || '',
            pageUrl: window.location.href
        };

        courses.push(course);

        console.log(`${index + 1}. ${title}`);
        console.log(`   Link: ${link || 'No direct link found'}`);
        console.log(`   Button: ${button.textContent.trim()}`);
        console.log('');
    });

    // Also look for course cards that might not have "Start Now" buttons
    const courseCards = document.querySelectorAll('.course, .lesson, .program, .module, [class*="course-item"], [class*="lesson-item"], [class*="card"]');
    courseCards.forEach((card, index) => {
        // Skip if we already processed this card
        if (courses.some(c => c.containerClass === card.className)) {
            return;
        }

        const link = card.querySelector('a[href]');
        const title = card.querySelector('h1, h2, h3, h4, .title, .course-title')?.textContent.trim();

        if (link && title) {
            courses.push({
                id: courses.length + 1,
                title: title,
                description: card.querySelector('.description, p')?.textContent.trim().substring(0, 200) || '',
                link: link.href,
                buttonText: 'Course Card',
                containerClass: card.className,
                pageUrl: window.location.href
            });
        }
    });

    // Remove duplicates based on links and titles
    const uniqueCourses = courses.filter((course, index, self) =>
        index === self.findIndex((c) =>
            (c.link && c.link === course.link) ||
            (c.title === course.title && c.title.length > 10)
        )
    );

    console.log('\n' + '='.repeat(80));
    console.log('📚 STAGE 1 COMPLETE - COURSES FOUND');
    console.log('='.repeat(80));
    console.log(`📊 Total unique courses/lessons: ${uniqueCourses.length}`);

    console.log('\n📋 COURSE LIST:');
    uniqueCourses.forEach((course, index) => {
        console.log(`${index + 1}. ${course.title}`);
        console.log(`   🔗 Link: ${course.link || 'No link'}`);
        console.log(`   📝 Description: ${course.description.substring(0, 100)}...`);
        console.log('');
    });

    // Create data for stage 2
    const scraperData = {
        stage: 'complete',
        courses: uniqueCourses,
        totalCourses: uniqueCourses.length,
        timestamp: new Date().toISOString(),
        libraryPage: window.location.href,
        nextStage: 'Run stage2_video_extractor.js'
    };

    // Auto-download the data
    const dataStr = JSON.stringify(scraperData, null, 2);
    const dataBlob = new Blob([dataStr], {type: 'application/json'});
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'grantcardone_courses.json';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    console.log('✅ Course data auto-downloaded as grantcardone_courses.json');
    console.log('🚀 Next step: Run python3 stage2_scraper.py to process these courses');

    // Also show data for manual copy
    console.log('\n📋 COPY THIS DATA FOR STAGE 2:');
    console.log('='.repeat(40));
    console.log(dataStr);

    return {
        courses: uniqueCourses,
        totalCourses: uniqueCourses.length,
        nextStep: 'Run python3 stage2_scraper.py'
    };
})();