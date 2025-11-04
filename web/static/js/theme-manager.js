// Theme Management System
// Provides comprehensive light/dark theme switching with system preference detection

// Alpine.js Component for Theme Management
document.addEventListener('alpine:init', () => {
    Alpine.data('themeManager', () => ({
        theme: 'system', // 'light', 'dark', or 'system'
        
        init() {
            // Load saved theme preference or default to system
            this.theme = localStorage.getItem('vip-panel-theme') || 'system';
            this.applyTheme();
            
            // Listen for system theme changes
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            mediaQuery.addEventListener('change', () => {
                if (this.theme === 'system') {
                    this.applyTheme();
                }
            });
        },

        setTheme(newTheme) {
            this.theme = newTheme;
            localStorage.setItem('vip-panel-theme', newTheme);
            this.applyTheme();
            
            // Trigger custom event for other components
            window.dispatchEvent(new CustomEvent('theme-changed', { 
                detail: { theme: newTheme, isDark: this.isDark() } 
            }));
        },

        applyTheme() {
            const isDark = this.isDark();
            const html = document.documentElement;
            
            if (isDark) {
                html.classList.add('dark');
            } else {
                html.classList.remove('dark');
            }
            
            // Update meta theme-color for mobile browsers
            const metaThemeColor = document.querySelector('meta[name="theme-color"]');
            if (metaThemeColor) {
                metaThemeColor.content = isDark ? '#0f172a' : '#ffffff';
            }
        },

        isDark() {
            if (this.theme === 'dark') return true;
            if (this.theme === 'light') return false;
            
            // System preference
            return window.matchMedia('(prefers-color-scheme: dark)').matches;
        },

        toggleTheme() {
            // Quick toggle between light and dark (ignoring system)
            const newTheme = this.isDark() ? 'light' : 'dark';
            this.setTheme(newTheme);
        }
    }));
});

// Utility functions for theme management
const ThemeUtils = {
    // Get current theme state
    getCurrentTheme() {
        return localStorage.getItem('vip-panel-theme') || 'system';
    },

    // Check if current theme is dark
    isDarkMode() {
        const theme = this.getCurrentTheme();
        if (theme === 'dark') return true;
        if (theme === 'light') return false;
        return window.matchMedia('(prefers-color-scheme: dark)').matches;
    },

    // Apply theme without Alpine.js (for early page load)
    applyThemeEarly() {
        const theme = this.getCurrentTheme();
        const isDark = theme === 'dark' || 
                      (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
        
        if (isDark) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    }
};

// Apply theme immediately to prevent flash of unstyled content
ThemeUtils.applyThemeEarly();

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { ThemeUtils };
}