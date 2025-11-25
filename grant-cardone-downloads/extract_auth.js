// Grant Cardone Auth Token Extractor
// Run this in Chrome DevTools console on the training site

(function() {
    console.log('🔑 Grant Cardone Auth Token Extractor');
    console.log('=====================================');

    const authData = {
        cookies: {},
        localStorage: {},
        sessionStorage: {},
        windowVars: {},
        headers: {},
        timestamp: new Date().toISOString()
    };

    // Extract all cookies
    document.cookie.split(';').forEach(cookie => {
        const [name, value] = cookie.trim().split('=');
        if (name && value) {
            authData.cookies[name] = decodeURIComponent(value);
        }
    });

    // Extract all localStorage
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        authData.localStorage[key] = localStorage.getItem(key);
    }

    // Extract all sessionStorage
    for (let i = 0; i < sessionStorage.length; i++) {
        const key = sessionStorage.key(i);
        authData.sessionStorage[key] = sessionStorage.getItem(key);
    }

    // Look for common auth variables in window
    const authKeys = [
        'token', 'authToken', 'accessToken', 'jwt', 'apiToken',
        'user', 'auth', 'currentUser', 'userData', 'session',
        'sessionToken', 'bearerToken', 'oauthToken'
    ];

    authKeys.forEach(key => {
        if (window[key] !== undefined) {
            authData.windowVars[key] = window[key];
        }
    });

    // Look for auth data in common libraries
    if (window.$ && window.$.ajaxSettings) {
        authData.headers['X-CSRF-TOKEN'] = window.$('meta[name="csrf-token"]').attr('content');
    }

    // Look for axios defaults
    if (window.axios && window.axios.defaults) {
        authData.headers['Authorization'] = window.axios.defaults.headers.common['Authorization'];
        authData.headers['X-CSRF-TOKEN'] = window.axios.defaults.headers.common['X-CSRF-TOKEN'];
    }

    // Check for jQuery AJAX headers
    if (window.jQuery && window.jQuery.ajaxSettings && window.jQuery.ajaxSettings.headers) {
        Object.assign(authData.headers, window.jQuery.ajaxSettings.headers);
    }

    // Look for meta tags
    document.querySelectorAll('meta').forEach(meta => {
        const name = meta.getAttribute('name') || meta.getAttribute('property');
        const content = meta.getAttribute('content');
        if (name && content && (name.includes('token') || name.includes('csrf'))) {
            authData.headers[name] = content;
        }
    });

    // Create Authorization header if we have tokens
    const possibleTokens = [
        authData.localStorage.token,
        authData.localStorage.authToken,
        authData.localStorage.accessToken,
        authData.windowVars.token,
        authData.windowVars.authToken,
        authData.windowVars.accessToken,
        authData.cookies.token,
        authData.cookies.authToken,
        authData.cookies.accessToken
    ].filter(Boolean);

    if (possibleTokens.length > 0) {
        const token = possibleTokens[0];
        if (!token.startsWith('Bearer ')) {
            authData.headers['Authorization'] = `Bearer ${token}`;
        } else {
            authData.headers['Authorization'] = token;
        }
    }

    // Display results
    console.log('\n📊 EXTRACTION RESULTS:');
    console.log('=======================');
    console.log(`🍪 Cookies: ${Object.keys(authData.cookies).length}`);
    console.log(`💾 LocalStorage: ${Object.keys(authData.localStorage).length}`);
    console.log(`🗄️ SessionStorage: ${Object.keys(authData.sessionStorage).length}`);
    console.log(`🪟 Window Variables: ${Object.keys(authData.windowVars).length}`);
    console.log(`📋 Headers: ${Object.keys(authData.headers).length}`);

    // Show important data
    console.log('\n🔑 AUTHENTICATION DATA:');
    console.log('========================');

    Object.keys(authData.headers).forEach(key => {
        console.log(`${key}: ${authData.headers[key]}`);
    });

    const importantCookies = Object.keys(authData.cookies).filter(key =>
        key.toLowerCase().includes('token') ||
        key.toLowerCase().includes('auth') ||
        key.toLowerCase().includes('session')
    );

    if (importantCookies.length > 0) {
        console.log('\n🍪 IMPORTANT COOKIES:');
        importantCookies.forEach(key => {
            console.log(`${key}: ${authData.cookies[key]}`);
        });
    }

    const importantStorage = Object.keys(authData.localStorage).filter(key =>
        key.toLowerCase().includes('token') ||
        key.toLowerCase().includes('auth') ||
        key.toLowerCase().includes('session') ||
        key.toLowerCase().includes('user')
    );

    if (importantStorage.length > 0) {
        console.log('\n💾 IMPORTANT LOCALSTORAGE:');
        importantStorage.forEach(key => {
            console.log(`${key}: ${authData.localStorage[key]}`);
        });
    }

    // Generate downloadable JSON
    console.log('\n📋 COPY THIS JSON:');
    console.log('==================');
    console.log(JSON.stringify(authData, null, 2));

    // Also save to file for download
    const dataStr = JSON.stringify(authData, null, 2);
    const dataBlob = new Blob([dataStr], {type: 'application/json'});
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'grantcardone_auth.json';
    link.click();
    URL.revokeObjectURL(url);

    console.log('\n✅ Auth data downloaded to grantcardone_auth.json');
    console.log('📁 Copy this file to the same directory as the downloader script');

    return authData;
})();