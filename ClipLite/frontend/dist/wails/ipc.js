window.wails = {
    Invoke: function(methodName, ...args) {
        return new Promise((resolve, reject) => {
            window.chrome.webview.postMessage({
                method: methodName,
                args: args
            });
            window.chrome.webview.addEventListener('message', function handler(event) {
                window.chrome.webview.removeEventListener('message', handler);
                if (event.data.error) {
                    reject(event.data.error);
                } else {
                    resolve(event.data.result);
                }
            });
        });
    }
};

window.go = {
    main: {
        App: {
            GetRecords: function(search) {
                return window.wails.Invoke('App.GetRecords', search);
            },
            GetFavorites: function() {
                return window.wails.Invoke('App.GetFavorites');
            },
            ToggleFavorite: function(id) {
                return window.wails.Invoke('App.ToggleFavorite', id);
            },
            DeleteRecord: function(id) {
                return window.wails.Invoke('App.DeleteRecord', id);
            },
            ClearAllRecords: function() {
                return window.wails.Invoke('App.ClearAllRecords');
            },
            GetConfig: function() {
                return window.wails.Invoke('App.GetConfig');
            },
            UpdateConfig: function(hotkey, saveDays, storagePath) {
                return window.wails.Invoke('App.UpdateConfig', hotkey, saveDays, storagePath);
            },
            CopyToClipboard: function(content) {
                return window.wails.Invoke('App.CopyToClipboard', content);
            }
        }
    }
};
