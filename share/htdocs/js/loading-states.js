/**
 * EntityDB Loading States
 * Reusable loading state components
 * Version: v2.29.0
 */

const LoadingStates = {
    components: {
        /**
         * Spinning loader component
         */
        'loading-spinner': {
            props: {
                message: {
                    type: String,
                    default: ''
                },
                size: {
                    type: String,
                    default: 'medium',
                    validator: value => ['small', 'medium', 'large'].includes(value)
                }
            },
            template: `
                <div class="loading-spinner" :class="'loading-spinner--' + size">
                    <div class="spinner">
                        <div class="spinner-circle"></div>
                    </div>
                    <p v-if="message" class="loading-message">{{ message }}</p>
                </div>
            `,
            mounted() {
                console.log('[LoadingSpinner] Mounted with message:', this.message);
            }
        },
        
        /**
         * Skeleton loader for content placeholders
         */
        'skeleton-loader': {
            props: {
                width: {
                    type: String,
                    default: '100%'
                },
                height: {
                    type: String,
                    default: '20px'
                },
                type: {
                    type: String,
                    default: 'text',
                    validator: value => ['text', 'title', 'avatar', 'button', 'image'].includes(value)
                },
                animated: {
                    type: Boolean,
                    default: true
                }
            },
            computed: {
                skeletonStyle() {
                    const styles = {
                        width: this.width,
                        height: this.height
                    };
                    
                    // Add type-specific styles
                    switch (this.type) {
                        case 'avatar':
                            styles.width = this.width || '40px';
                            styles.height = this.height || '40px';
                            styles.borderRadius = '50%';
                            break;
                        case 'title':
                            styles.height = this.height || '32px';
                            break;
                        case 'button':
                            styles.width = this.width || '120px';
                            styles.height = this.height || '36px';
                            styles.borderRadius = '6px';
                            break;
                        case 'image':
                            styles.width = this.width || '100%';
                            styles.height = this.height || '200px';
                            styles.borderRadius = '8px';
                            break;
                    }
                    
                    return styles;
                }
            },
            template: `
                <div 
                    class="skeleton-loader" 
                    :class="{ 'skeleton-animated': animated }"
                    :style="skeletonStyle"
                >
                    <div class="skeleton-shimmer" v-if="animated"></div>
                </div>
            `
        },
        
        /**
         * Progress bar component
         */
        'progress-bar': {
            props: {
                value: {
                    type: Number,
                    required: true,
                    validator: value => value >= 0 && value <= 100
                },
                message: {
                    type: String,
                    default: ''
                },
                showPercentage: {
                    type: Boolean,
                    default: true
                },
                color: {
                    type: String,
                    default: '#3498db'
                }
            },
            template: `
                <div class="progress-container">
                    <div class="progress-header" v-if="message || showPercentage">
                        <span v-if="message" class="progress-message">{{ message }}</span>
                        <span v-if="showPercentage" class="progress-percentage">{{ Math.round(value) }}%</span>
                    </div>
                    <div class="progress-bar">
                        <div 
                            class="progress-fill" 
                            :style="{ width: value + '%', backgroundColor: color }"
                        ></div>
                    </div>
                </div>
            `
        },
        
        /**
         * Loading overlay component
         */
        'loading-overlay': {
            props: {
                visible: {
                    type: Boolean,
                    default: true
                },
                message: {
                    type: String,
                    default: 'Loading...'
                },
                fullscreen: {
                    type: Boolean,
                    default: false
                }
            },
            template: `
                <transition name="fade">
                    <div 
                        v-if="visible" 
                        class="loading-overlay" 
                        :class="{ 'loading-overlay--fullscreen': fullscreen }"
                    >
                        <div class="loading-overlay-content">
                            <loading-spinner :message="message" size="large" />
                        </div>
                    </div>
                </transition>
            `
        },
        
        /**
         * Content loader with multiple skeletons
         */
        'content-loader': {
            props: {
                type: {
                    type: String,
                    default: 'list',
                    validator: value => ['list', 'card', 'table', 'form'].includes(value)
                },
                count: {
                    type: Number,
                    default: 3
                }
            },
            computed: {
                skeletonLayout() {
                    const layouts = {
                        list: [
                            { type: 'text', width: '60%', height: '16px' },
                            { type: 'text', width: '40%', height: '14px', marginTop: '8px' }
                        ],
                        card: [
                            { type: 'image', width: '100%', height: '150px' },
                            { type: 'title', width: '80%', marginTop: '16px' },
                            { type: 'text', width: '100%', marginTop: '12px' },
                            { type: 'text', width: '60%', marginTop: '8px' }
                        ],
                        table: [
                            { type: 'text', width: '20%', height: '16px' },
                            { type: 'text', width: '30%', height: '16px', marginLeft: '20px' },
                            { type: 'text', width: '25%', height: '16px', marginLeft: '20px' },
                            { type: 'text', width: '15%', height: '16px', marginLeft: '20px' }
                        ],
                        form: [
                            { type: 'text', width: '30%', height: '14px' },
                            { type: 'text', width: '100%', height: '40px', marginTop: '8px' },
                            { type: 'text', width: '30%', height: '14px', marginTop: '16px' },
                            { type: 'text', width: '100%', height: '40px', marginTop: '8px' }
                        ]
                    };
                    return layouts[this.type] || layouts.list;
                }
            },
            template: `
                <div class="content-loader">
                    <div 
                        v-for="i in count" 
                        :key="i" 
                        class="skeleton-group"
                        :style="{ marginBottom: '24px' }"
                    >
                        <skeleton-loader
                            v-for="(skeleton, index) in skeletonLayout"
                            :key="index"
                            :type="skeleton.type"
                            :width="skeleton.width"
                            :height="skeleton.height"
                            :style="{
                                marginTop: skeleton.marginTop || '0',
                                marginLeft: skeleton.marginLeft || '0',
                                display: skeleton.marginLeft ? 'inline-block' : 'block'
                            }"
                        />
                    </div>
                </div>
            `
        }
    }
};

// Add CSS styles
const loadingStyles = document.createElement('style');
loadingStyles.textContent = `
/* Loading Spinner */
.loading-spinner {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 20px;
}

.loading-spinner--small .spinner {
    width: 24px;
    height: 24px;
}

.loading-spinner--medium .spinner {
    width: 40px;
    height: 40px;
}

.loading-spinner--large .spinner {
    width: 56px;
    height: 56px;
}

.spinner {
    position: relative;
    display: inline-block;
}

.spinner-circle {
    width: 100%;
    height: 100%;
    border: 3px solid rgba(52, 152, 219, 0.2);
    border-top-color: #3498db;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

.loading-message {
    margin-top: 12px;
    color: #6c757d;
    font-size: 14px;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

/* Skeleton Loader */
.skeleton-loader {
    background: #e9ecef;
    position: relative;
    overflow: hidden;
}

body.dark-mode .skeleton-loader {
    background: #34495e;
}

.skeleton-animated::after {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(
        90deg,
        transparent,
        rgba(255, 255, 255, 0.4),
        transparent
    );
    animation: shimmer 1.5s infinite;
}

body.dark-mode .skeleton-animated::after {
    background: linear-gradient(
        90deg,
        transparent,
        rgba(255, 255, 255, 0.1),
        transparent
    );
}

@keyframes shimmer {
    to { left: 100%; }
}

/* Progress Bar */
.progress-container {
    width: 100%;
}

.progress-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
}

.progress-message {
    color: #6c757d;
    font-size: 14px;
}

.progress-percentage {
    color: #6c757d;
    font-size: 14px;
    font-weight: 500;
}

.progress-bar {
    width: 100%;
    height: 8px;
    background: #e9ecef;
    border-radius: 4px;
    overflow: hidden;
}

body.dark-mode .progress-bar {
    background: #34495e;
}

.progress-fill {
    height: 100%;
    transition: width 0.3s ease;
    border-radius: 4px;
}

/* Loading Overlay */
.loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(255, 255, 255, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

body.dark-mode .loading-overlay {
    background: rgba(26, 29, 33, 0.9);
}

.loading-overlay--fullscreen {
    position: fixed;
}

.loading-overlay-content {
    text-align: center;
}

/* Transitions */
.fade-enter-active, .fade-leave-active {
    transition: opacity 0.3s;
}

.fade-enter-from, .fade-leave-to {
    opacity: 0;
}

/* Content Loader */
.skeleton-group {
    padding: 16px;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

body.dark-mode .skeleton-group {
    background: #2c3e50;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
`;
document.head.appendChild(loadingStyles);

// Export for use in other modules
window.LoadingStates = LoadingStates;