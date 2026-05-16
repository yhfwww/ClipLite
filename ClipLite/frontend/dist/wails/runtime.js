window.runtime = {
    WindowHide: function() {
        return window.wailsInvoke('runtime.WindowHide');
    },
    WindowShow: function() {
        return window.wailsInvoke('runtime.WindowShow');
    },
    WindowMinimise: function() {
        return window.wailsInvoke('runtime.WindowMinimise');
    },
    WindowMaximise: function() {
        return window.wailsInvoke('runtime.WindowMaximise');
    },
    WindowUnmaximise: function() {
        return window.wailsInvoke('runtime.WindowUnmaximise');
    },
    WindowToggleMaximise: function() {
        return window.wailsInvoke('runtime.WindowToggleMaximise');
    },
    WindowClose: function() {
        return window.wailsInvoke('runtime.WindowClose');
    },
    OpenDirectoryDialog: function(title) {
        return window.wailsInvoke('runtime.OpenDirectoryDialog', title);
    },
    OpenFileDialog: function(title) {
        return window.wailsInvoke('runtime.OpenFileDialog', title);
    },
    SaveFileDialog: function(title) {
        return window.wailsInvoke('runtime.SaveFileDialog', title);
    }
};

function wailsInvoke(methodName, ...args) {
    return new Promise((resolve, reject) => {
        try {
            window.chrome.webview.postMessage({
                method: methodName,
                args: args
            });
            const handler = function(event) {
                window.chrome.webview.removeEventListener('message', handler);
                if (event.data && event.data.error) {
                    reject(new Error(event.data.error));
                } else {
                    resolve(event.data ? event.data.result : null);
                }
            };
            window.chrome.webview.addEventListener('message', handler);
        } catch (e) {
            reject(e);
        }
    });
}
