// Debug Console - EntityDB v2.31.0
// Enhanced console debugging and issue resolution tools

class DebugConsole {
    constructor() {
        this.commands = new Map();
        this.debugMode = false;
        this.activeSession = null;
        this.commandHistory = [];
        this.maxHistory = 100;
        
        this.initializeCommands();
        this.setupConsoleEnhancements();
    }
    
    // Initialize debug commands
    initializeCommands() {
        // Error management commands
        this.addCommand('errors.list', (filter = {}) => {
            return window.ErrorHandler?.getErrors(filter) || [];
        }, 'List errors with optional filtering');
        
        this.addCommand('errors.clear', () => {
            window.ErrorHandler?.clearErrors();
            return 'All errors cleared';
        }, 'Clear all error history');
        
        this.addCommand('errors.export', () => {
            const data = window.ErrorHandler?.exportErrorData();
            console.log('Error data:', data);
            return data;
        }, 'Export error data');
        
        this.addCommand('errors.simulate', (type, message, category = 'test') => {
            window.ErrorHandler?.createError(type, message, category, 'medium');
            return `Simulated error: ${type}`;
        }, 'Simulate an error for testing');
        
        // Request debugging commands
        this.addCommand('requests.list', (filter = {}) => {
            return window.RequestDebugger?.getLogs(filter) || [];
        }, 'List API requests with optional filtering');
        
        this.addCommand('requests.timeline', (requestId) => {
            return window.RequestDebugger?.getRequestTimeline(requestId) || [];
        }, 'Get request timeline by ID');
        
        this.addCommand('requests.stats', () => {
            return window.RequestDebugger?.getStats() || {};
        }, 'Get request statistics');
        
        this.addCommand('requests.clear', () => {
            window.RequestDebugger?.clearLogs();
            return 'Request history cleared';
        }, 'Clear request history');
        
        // UI health commands
        this.addCommand('ui.health', () => {
            return window.UIResilience?.performHealthCheck() || {};
        }, 'Perform UI health check');
        
        this.addCommand('ui.components', () => {
            return window.UIResilience?.getComponentMetrics() || {};
        }, 'Get component health metrics');
        
        this.addCommand('ui.recover', (componentName) => {
            if (!componentName) {
                return 'Usage: ui.recover("componentName")';
            }
            window.UIResilience?.resetComponentState(componentName);
            return `Reset state for component: ${componentName}`;
        }, 'Reset component state');
        
        this.addCommand('ui.validate', () => {
            const app = window.entitydbApp;
            if (!app) return 'App not available';
            
            const validation = {
                app: !!app,
                authenticated: app.isAuthenticated,
                activeTab: app.activeTab,
                tabs: app.tabs?.length || 0,
                darkMode: app.isDarkMode,
                systemStats: !!app.systemStats,
                errors: Object.keys(app.errorData || {}).length
            };
            
            return validation;
        }, 'Validate UI state');
        
        // API debugging commands
        this.addCommand('api.test', async (endpoint = '/health') => {
            try {
                const response = await fetch(endpoint);
                const data = await response.json();
                return { status: response.status, data };
            } catch (error) {
                return { error: error.message };
            }
        }, 'Test API endpoint');
        
        this.addCommand('api.auth', () => {
            const app = window.entitydbApp;
            if (!app) return 'App not available';
            
            return {
                authenticated: app.isAuthenticated,
                user: app.currentUser,
                token: app.sessionToken ? 'Present' : 'Missing',
                dataset: app.currentDataset
            };
        }, 'Check authentication status');
        
        // Performance debugging
        this.addCommand('perf.memory', () => {
            if (!performance.memory) return 'Memory API not available';
            
            const memory = performance.memory;
            return {
                used: Math.round(memory.usedJSHeapSize / 1024 / 1024) + 'MB',
                total: Math.round(memory.totalJSHeapSize / 1024 / 1024) + 'MB',
                limit: Math.round(memory.jsHeapSizeLimit / 1024 / 1024) + 'MB',
                usage: Math.round((memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100) + '%'
            };
        }, 'Get memory usage statistics');
        
        this.addCommand('perf.gc', () => {
            if (window.gc) {
                window.gc();
                return 'Garbage collection triggered';
            }
            return 'Garbage collection not available';
        }, 'Trigger garbage collection');
        
        // Tab debugging
        this.addCommand('tabs.validate', () => {
            return window.UIResilience?.tabValidator?.validate() || ['Validator not available'];
        }, 'Validate tab structure');
        
        this.addCommand('tabs.switch', (tabId) => {
            const app = window.entitydbApp;
            if (!app) return 'App not available';
            if (!tabId) return 'Usage: tabs.switch("tabId")';
            
            app.switchTab(tabId);
            return `Switched to tab: ${tabId}`;
        }, 'Switch to specific tab');
        
        this.addCommand('tabs.list', () => {
            const app = window.entitydbApp;
            if (!app) return 'App not available';
            
            return app.tabs?.map(tab => ({
                id: tab.id,
                name: tab.name,
                active: tab.id === app.activeTab
            })) || [];
        }, 'List all available tabs');
        
        // Debug session commands
        this.addCommand('debug.start', () => {
            this.debugMode = true;
            this.activeSession = {
                started: new Date().toISOString(),
                commands: [],
                errors: []
            };
            
            // Enable verbose logging
            window.ErrorHandler?.enableDebug();
            window.RequestDebugger?.setDebugLevel('verbose');
            
            console.log('🐛 Debug session started');
            return 'Debug session active';
        }, 'Start debug session');
        
        this.addCommand('debug.stop', () => {
            this.debugMode = false;
            
            // Disable verbose logging
            window.ErrorHandler?.disableDebug();
            window.RequestDebugger?.setDebugLevel('info');
            
            const session = this.activeSession;
            this.activeSession = null;
            
            console.log('🐛 Debug session ended');
            return session;
        }, 'Stop debug session');
        
        this.addCommand('debug.status', () => {
            return {
                active: this.debugMode,
                session: this.activeSession,
                commandsAvailable: this.commands.size,
                historyLength: this.commandHistory.length
            };
        }, 'Get debug session status');
        
        // Utility commands
        this.addCommand('help', (category) => {
            if (category) {
                const filtered = Array.from(this.commands.entries())
                    .filter(([name]) => name.startsWith(category));
                return filtered.map(([name, cmd]) => ({
                    command: name,
                    description: cmd.description
                }));
            }
            
            const categories = {};
            for (const [name, cmd] of this.commands) {
                const category = name.split('.')[0];
                if (!categories[category]) categories[category] = [];
                categories[category].push({ command: name, description: cmd.description });
            }
            
            return categories;
        }, 'Show available commands (optionally filtered by category)');
        
        this.addCommand('history', () => {
            return this.commandHistory.slice(-20); // Last 20 commands
        }, 'Show command history');
    }
    
    // Add a debug command
    addCommand(name, handler, description) {
        this.commands.set(name, {
            handler,
            description,
            lastUsed: null,
            useCount: 0
        });
    }
    
    // Execute a debug command
    async executeCommand(commandName, ...args) {
        const command = this.commands.get(commandName);
        if (!command) {
            return `Command '${commandName}' not found. Use help() for available commands.`;
        }
        
        try {
            // Record command usage
            command.lastUsed = new Date().toISOString();
            command.useCount++;
            
            // Add to history
            this.commandHistory.push({
                command: commandName,
                args,
                timestamp: new Date().toISOString()
            });
            
            // Limit history size
            if (this.commandHistory.length > this.maxHistory) {
                this.commandHistory = this.commandHistory.slice(-this.maxHistory);
            }
            
            // Execute command
            const result = await command.handler(...args);
            
            // Log to session if active
            if (this.activeSession) {
                this.activeSession.commands.push({
                    command: commandName,
                    args,
                    result: typeof result === 'object' ? JSON.stringify(result) : result,
                    timestamp: new Date().toISOString()
                });
            }
            
            return result;
        } catch (error) {
            const errorMsg = `Error executing '${commandName}': ${error.message}`;
            
            // Log error to session if active
            if (this.activeSession) {
                this.activeSession.errors.push({
                    command: commandName,
                    error: error.message,
                    stack: error.stack,
                    timestamp: new Date().toISOString()
                });
            }
            
            return errorMsg;
        }
    }
    
    // Setup console enhancements
    setupConsoleEnhancements() {
        // Add global debug object
        if (typeof window !== 'undefined') {
            window.debug = this.createDebugProxy();
            
            // Add convenience methods
            window.dbg = window.debug; // Short alias
            
            // Add individual command shortcuts
            window.errors = {
                list: (filter) => this.executeCommand('errors.list', filter),
                clear: () => this.executeCommand('errors.clear'),
                export: () => this.executeCommand('errors.export'),
                simulate: (type, msg, cat) => this.executeCommand('errors.simulate', type, msg, cat)
            };
            
            window.requests = {
                list: (filter) => this.executeCommand('requests.list', filter),
                timeline: (id) => this.executeCommand('requests.timeline', id),
                stats: () => this.executeCommand('requests.stats'),
                clear: () => this.executeCommand('requests.clear')
            };
            
            window.ui = {
                health: () => this.executeCommand('ui.health'),
                components: () => this.executeCommand('ui.components'),
                recover: (name) => this.executeCommand('ui.recover', name),
                validate: () => this.executeCommand('ui.validate')
            };
            
            window.tabs = {
                validate: () => this.executeCommand('tabs.validate'),
                switch: (id) => this.executeCommand('tabs.switch', id),
                list: () => this.executeCommand('tabs.list')
            };
        }
        
        // Override console methods for session logging
        this.setupConsoleLogging();
    }
    
    // Create debug proxy for command execution
    createDebugProxy() {
        return new Proxy({}, {
            get: (target, property) => {
                if (property === 'help') {
                    return (category) => this.executeCommand('help', category);
                }
                
                if (this.commands.has(property)) {
                    return (...args) => this.executeCommand(property, ...args);
                }
                
                // Try command with dots
                const commandName = property.toString();
                if (this.commands.has(commandName)) {
                    return (...args) => this.executeCommand(commandName, ...args);
                }
                
                return `Command '${property}' not found. Use debug.help() for available commands.`;
            }
        });
    }
    
    // Setup console logging for debug sessions
    setupConsoleLogging() {
        if (!window.console || this.consolePatched) return;
        
        const originalError = console.error;
        const originalWarn = console.warn;
        const originalLog = console.log;
        
        console.error = (...args) => {
            if (this.activeSession) {
                this.activeSession.errors.push({
                    type: 'console.error',
                    message: args.join(' '),
                    timestamp: new Date().toISOString()
                });
            }
            originalError.apply(console, args);
        };
        
        console.warn = (...args) => {
            if (this.activeSession) {
                this.activeSession.commands.push({
                    type: 'console.warn',
                    message: args.join(' '),
                    timestamp: new Date().toISOString()
                });
            }
            originalWarn.apply(console, args);
        };
        
        this.consolePatched = true;
    }
    
    // Quick issue diagnosis
    async diagnoseIssue(description) {
        console.log(`🔍 Diagnosing: ${description}`);
        
        const diagnosis = {
            timestamp: new Date().toISOString(),
            description,
            checks: {}
        };
        
        // UI Health
        diagnosis.checks.uiHealth = await this.executeCommand('ui.health');
        diagnosis.checks.uiValidation = await this.executeCommand('ui.validate');
        
        // Error Analysis
        const recentErrors = await this.executeCommand('errors.list', {
            since: new Date(Date.now() - 300000).toISOString() // Last 5 minutes
        });
        diagnosis.checks.recentErrors = recentErrors.length;
        diagnosis.checks.errorCategories = [...new Set(recentErrors.map(e => e.category))];
        
        // Request Analysis
        diagnosis.checks.requestStats = await this.executeCommand('requests.stats');
        
        // Memory Check
        diagnosis.checks.memory = await this.executeCommand('perf.memory');
        
        // Tab Structure
        diagnosis.checks.tabIssues = await this.executeCommand('tabs.validate');
        
        // Authentication
        diagnosis.checks.auth = await this.executeCommand('api.auth');
        
        console.log('🔍 Diagnosis complete:', diagnosis);
        return diagnosis;
    }
    
    // Auto-fix common issues
    async autoFix() {
        console.log('🔧 Attempting auto-fix...');
        
        const fixes = [];
        
        // Clear old errors
        const errorCount = (await this.executeCommand('errors.list')).length;
        if (errorCount > 50) {
            await this.executeCommand('errors.clear');
            fixes.push(`Cleared ${errorCount} old errors`);
        }
        
        // Clear old requests
        const requestCount = (await this.executeCommand('requests.list')).length;
        if (requestCount > 100) {
            await this.executeCommand('requests.clear');
            fixes.push(`Cleared ${requestCount} old requests`);
        }
        
        // Garbage collection
        const gcResult = await this.executeCommand('perf.gc');
        if (gcResult.includes('triggered')) {
            fixes.push('Triggered garbage collection');
        }
        
        // UI health check
        const health = await this.executeCommand('ui.health');
        if (health.status !== 'healthy') {
            fixes.push(`UI health status: ${health.status}`);
        }
        
        console.log('🔧 Auto-fix complete:', fixes);
        return fixes;
    }
    
    // Get diagnostic report
    async getReport() {
        const report = {
            timestamp: new Date().toISOString(),
            system: await this.executeCommand('ui.validate'),
            health: await this.executeCommand('ui.health'),
            errors: (await this.executeCommand('errors.list')).slice(0, 10),
            requests: await this.executeCommand('requests.stats'),
            memory: await this.executeCommand('perf.memory'),
            auth: await this.executeCommand('api.auth'),
            tabs: await this.executeCommand('tabs.list'),
            commandHistory: this.commandHistory.slice(-10)
        };
        
        console.log('📊 System Report:', report);
        return report;
    }
}

// Global debug console instance
window.DebugConsole = window.DebugConsole || new DebugConsole();

// Quick start message
if (typeof window !== 'undefined' && window.console) {
    console.log(`
🐛 EntityDB Debug Console Ready!

Quick commands:
• debug.help()           - Show all commands
• errors.list()          - List recent errors  
• ui.health()           - Check UI health
• requests.stats()       - API request stats
• tabs.validate()        - Check tab structure
• debug.start()         - Start debug session
• diagnose("issue")     - Diagnose problems
• autoFix()             - Auto-fix common issues

Full report: getReport()
    `);
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DebugConsole;
}