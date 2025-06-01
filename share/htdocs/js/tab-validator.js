/**
 * Tab Structure Validator
 * Ensures tabs use x-if templates and not x-show directives
 */

class TabValidator {
    constructor() {
        this.errors = [];
        this.warnings = [];
    }

    /**
     * Validate tab structure on page load
     */
    validateTabs() {
        console.log('🔍 Validating tab structure...');
        
        // Check for any x-show on tab-content elements
        const invalidTabs = document.querySelectorAll('[x-show*="activeTab"].tab-content');
        if (invalidTabs.length > 0) {
            this.errors.push(`Found ${invalidTabs.length} tabs using x-show instead of x-if templates`);
            invalidTabs.forEach(tab => {
                const condition = tab.getAttribute('x-show');
                console.error(`❌ Invalid tab: ${condition} - should use x-if with template`);
            });
        }

        // Check that all templates have x-if for activeTab
        const templates = document.querySelectorAll('template[x-if*="activeTab"]');
        console.log(`✅ Found ${templates.length} valid tab templates`);

        // Check for nested tab-content (common mistake)
        const nestedTabs = document.querySelectorAll('.tab-content .tab-content');
        if (nestedTabs.length > 0) {
            this.errors.push(`Found ${nestedTabs.length} nested tab-content elements`);
        }

        // Check main-content structure
        const mainContent = document.querySelector('.main-content');
        if (mainContent) {
            const computedStyle = window.getComputedStyle(mainContent);
            
            // Verify flex layout
            if (computedStyle.display !== 'flex') {
                this.warnings.push('main-content should have display: flex');
            }
            
            // Check for problematic overflow
            if (computedStyle.overflow === 'visible') {
                this.warnings.push('main-content should have overflow: hidden to contain tabs');
            }
        }

        // Report results
        if (this.errors.length > 0) {
            console.error('🚨 Tab validation failed:', this.errors);
            this.showErrorBanner();
        } else {
            console.log('✅ Tab structure validation passed!');
        }

        if (this.warnings.length > 0) {
            console.warn('⚠️ Tab validation warnings:', this.warnings);
        }

        return this.errors.length === 0;
    }

    /**
     * Show error banner to developers
     */
    showErrorBanner() {
        const banner = document.createElement('div');
        banner.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            background: #ff4444;
            color: white;
            padding: 10px;
            z-index: 10000;
            text-align: center;
            font-weight: bold;
        `;
        banner.innerHTML = `
            ⚠️ Tab Structure Error: ${this.errors.join(', ')} 
            <button onclick="this.parentElement.remove()" style="margin-left: 10px;">Dismiss</button>
        `;
        document.body.prepend(banner);
    }

    /**
     * Auto-fix common issues (development mode only)
     */
    autoFix() {
        console.log('🔧 Attempting auto-fix...');
        
        // Find all x-show tabs
        const invalidTabs = document.querySelectorAll('[x-show*="activeTab"].tab-content');
        invalidTabs.forEach(tab => {
            const condition = tab.getAttribute('x-show');
            console.warn(`Would convert: ${condition} to x-if template structure`);
            // In production, we don't actually modify the DOM
            // This is just for logging what needs to be fixed
        });
    }
}

// Run validation when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        const validator = new TabValidator();
        validator.validateTabs();
    });
} else {
    const validator = new TabValidator();
    validator.validateTabs();
}

// Export for use in other scripts
window.TabValidator = TabValidator;