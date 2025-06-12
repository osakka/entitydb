/**
 * EntityDB Component Testing Framework
 * Lightweight testing framework for UI components
 */

class TestFramework {
    constructor() {
        this.tests = new Map();
        this.suites = new Map();
        this.mocks = new Map();
        this.fixtures = new Map();
        this.results = [];
        this.isRunning = false;
        this.reporter = new TestReporter();
    }

    /**
     * Define a test suite
     */
    describe(name, setupFn) {
        const suite = new TestSuite(name, this);
        this.suites.set(name, suite);
        
        // Execute setup function with suite context
        setupFn.call(suite, suite);
        
        return suite;
    }

    /**
     * Register a test
     */
    it(description, testFn, options = {}) {
        const test = new Test(description, testFn, options);
        this.tests.set(test.id, test);
        return test;
    }

    /**
     * Create a mock object
     */
    mock(name, implementation = {}) {
        const mock = new Mock(name, implementation);
        this.mocks.set(name, mock);
        return mock;
    }

    /**
     * Create a test fixture
     */
    fixture(name, data) {
        this.fixtures.set(name, data);
        return data;
    }

    /**
     * Get a fixture
     */
    getFixture(name) {
        return this.fixtures.get(name);
    }

    /**
     * Run all tests
     */
    async runAll() {
        if (this.isRunning) {
            throw new Error('Tests are already running');
        }

        this.isRunning = true;
        this.results = [];
        
        try {
            this.reporter.onStart(this.tests.size);
            
            // Run test suites
            for (const [name, suite] of this.suites) {
                await this.runSuite(suite);
            }
            
            // Run standalone tests
            for (const [id, test] of this.tests) {
                if (!test.suite) {
                    await this.runTest(test);
                }
            }
            
            this.reporter.onComplete(this.results);
            return this.results;
        } finally {
            this.isRunning = false;
        }
    }

    /**
     * Run a specific test suite
     */
    async runSuite(suite) {
        this.reporter.onSuiteStart(suite.name);
        
        try {
            // Run before hooks
            for (const hook of suite.beforeHooks) {
                await hook();
            }
            
            // Run tests
            for (const test of suite.tests) {
                await this.runTest(test, suite);
            }
            
            // Run after hooks
            for (const hook of suite.afterHooks) {
                await hook();
            }
            
            this.reporter.onSuiteEnd(suite.name);
        } catch (error) {
            this.reporter.onSuiteError(suite.name, error);
        }
    }

    /**
     * Run a single test
     */
    async runTest(test, suite = null) {
        const startTime = performance.now();
        const result = {
            id: test.id,
            description: test.description,
            suite: suite ? suite.name : null,
            status: 'pending',
            duration: 0,
            error: null,
            assertions: []
        };

        this.reporter.onTestStart(test.description);

        try {
            // Setup test environment
            const testEnv = this.createTestEnvironment(test, suite);
            
            // Run test
            await test.fn.call(testEnv, testEnv);
            
            result.status = 'passed';
            result.assertions = testEnv.assertions;
        } catch (error) {
            result.status = 'failed';
            result.error = error;
        } finally {
            result.duration = performance.now() - startTime;
            this.results.push(result);
            this.reporter.onTestEnd(result);
        }

        return result;
    }

    /**
     * Create test environment
     */
    createTestEnvironment(test, suite) {
        const assertions = [];
        
        const testEnv = {
            // Assertion methods
            expect: (actual) => new Expectation(actual, assertions),
            assert: new AssertionLibrary(assertions),
            
            // Mock utilities
            mock: (name) => this.mocks.get(name),
            createMock: (implementation) => new Mock('temp', implementation),
            
            // Fixture utilities
            fixture: (name) => this.getFixture(name),
            
            // DOM utilities
            dom: new DOMTestUtils(),
            
            // Component utilities
            component: new ComponentTestUtils(),
            
            // Async utilities
            waitFor: this.waitFor.bind(this),
            timeout: this.timeout.bind(this),
            
            // Test metadata
            test,
            suite,
            assertions
        };

        return testEnv;
    }

    /**
     * Wait for a condition to be true
     */
    async waitFor(condition, timeout = 5000, interval = 100) {
        const startTime = Date.now();
        
        while (Date.now() - startTime < timeout) {
            if (await condition()) {
                return true;
            }
            await this.delay(interval);
        }
        
        throw new Error(`waitFor timeout after ${timeout}ms`);
    }

    /**
     * Add timeout to promise
     */
    timeout(promise, ms) {
        return Promise.race([
            promise,
            new Promise((_, reject) => 
                setTimeout(() => reject(new Error(`Timeout after ${ms}ms`)), ms)
            )
        ]);
    }

    /**
     * Delay utility
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * Get test results summary
     */
    getSummary() {
        const total = this.results.length;
        const passed = this.results.filter(r => r.status === 'passed').length;
        const failed = this.results.filter(r => r.status === 'failed').length;
        const duration = this.results.reduce((sum, r) => sum + r.duration, 0);
        
        return {
            total,
            passed,
            failed,
            successRate: total > 0 ? (passed / total * 100).toFixed(2) : 0,
            duration: Math.round(duration)
        };
    }
}

/**
 * Test Suite class
 */
class TestSuite {
    constructor(name, framework) {
        this.name = name;
        this.framework = framework;
        this.tests = [];
        this.beforeHooks = [];
        this.afterHooks = [];
        this.beforeEachHooks = [];
        this.afterEachHooks = [];
    }

    it(description, testFn, options = {}) {
        const test = new Test(description, testFn, { ...options, suite: this.name });
        test.suite = this;
        this.tests.push(test);
        return test;
    }

    before(hookFn) {
        this.beforeHooks.push(hookFn);
    }

    after(hookFn) {
        this.afterHooks.push(hookFn);
    }

    beforeEach(hookFn) {
        this.beforeEachHooks.push(hookFn);
    }

    afterEach(hookFn) {
        this.afterEachHooks.push(hookFn);
    }
}

/**
 * Individual test class
 */
class Test {
    constructor(description, fn, options = {}) {
        this.id = Date.now() + Math.random();
        this.description = description;
        this.fn = fn;
        this.options = options;
        this.suite = null;
    }
}

/**
 * Expectation class for fluent assertions
 */
class Expectation {
    constructor(actual, assertions) {
        this.actual = actual;
        this.assertions = assertions;
        this.negated = false;
    }

    get not() {
        this.negated = !this.negated;
        return this;
    }

    toBe(expected) {
        const passed = this.negated ? this.actual !== expected : this.actual === expected;
        this.addAssertion('toBe', { expected }, passed);
        return this;
    }

    toEqual(expected) {
        const passed = this.negated ? 
            !this.deepEqual(this.actual, expected) : 
            this.deepEqual(this.actual, expected);
        this.addAssertion('toEqual', { expected }, passed);
        return this;
    }

    toContain(expected) {
        const passed = this.negated ? 
            !this.actual.includes(expected) : 
            this.actual.includes(expected);
        this.addAssertion('toContain', { expected }, passed);
        return this;
    }

    toBeGreaterThan(expected) {
        const passed = this.negated ? this.actual <= expected : this.actual > expected;
        this.addAssertion('toBeGreaterThan', { expected }, passed);
        return this;
    }

    toBeTruthy() {
        const passed = this.negated ? !this.actual : !!this.actual;
        this.addAssertion('toBeTruthy', {}, passed);
        return this;
    }

    toBeFalsy() {
        const passed = this.negated ? !!this.actual : !this.actual;
        this.addAssertion('toBeFalsy', {}, passed);
        return this;
    }

    toThrow(expected = null) {
        let passed = false;
        let error = null;
        
        try {
            this.actual();
        } catch (e) {
            error = e;
            passed = true;
            if (expected && e.message !== expected) {
                passed = false;
            }
        }
        
        passed = this.negated ? !passed : passed;
        this.addAssertion('toThrow', { expected, error }, passed);
        return this;
    }

    addAssertion(matcher, args, passed) {
        this.assertions.push({
            matcher,
            args,
            passed,
            actual: this.actual,
            negated: this.negated
        });

        if (!passed) {
            const message = this.negated ? 
                `Expected ${JSON.stringify(this.actual)} NOT to ${matcher}` :
                `Expected ${JSON.stringify(this.actual)} to ${matcher}`;
            throw new Error(message);
        }
    }

    deepEqual(a, b) {
        if (a === b) return true;
        if (a instanceof Date && b instanceof Date) return a.getTime() === b.getTime();
        if (!a || !b || (typeof a !== 'object' && typeof b !== 'object')) return a === b;
        if (a === null || a === undefined || b === null || b === undefined) return false;
        if (a.prototype !== b.prototype) return false;
        
        const keys = Object.keys(a);
        if (keys.length !== Object.keys(b).length) return false;
        
        return keys.every(k => this.deepEqual(a[k], b[k]));
    }
}

/**
 * Mock class
 */
class Mock {
    constructor(name, implementation = {}) {
        this.name = name;
        this.implementation = implementation;
        this.calls = [];
        
        return new Proxy(this, {
            get(target, prop) {
                if (prop in target) {
                    return target[prop];
                }
                
                if (prop in implementation) {
                    return (...args) => {
                        target.calls.push({ method: prop, args, timestamp: Date.now() });
                        return implementation[prop](...args);
                    };
                }
                
                return (...args) => {
                    target.calls.push({ method: prop, args, timestamp: Date.now() });
                    return undefined;
                };
            }
        });
    }

    reset() {
        this.calls = [];
    }

    wasCalledWith(method, ...args) {
        return this.calls.some(call => 
            call.method === method && 
            this.deepEqual(call.args, args)
        );
    }

    getCallCount(method) {
        return this.calls.filter(call => call.method === method).length;
    }
}

/**
 * DOM Test Utilities
 */
class DOMTestUtils {
    createElement(tag, attributes = {}, children = []) {
        const element = document.createElement(tag);
        
        Object.entries(attributes).forEach(([key, value]) => {
            element.setAttribute(key, value);
        });
        
        children.forEach(child => {
            if (typeof child === 'string') {
                element.appendChild(document.createTextNode(child));
            } else {
                element.appendChild(child);
            }
        });
        
        return element;
    }

    render(html) {
        const container = document.createElement('div');
        container.innerHTML = html;
        document.body.appendChild(container);
        return container;
    }

    cleanup(element) {
        if (element && element.parentNode) {
            element.parentNode.removeChild(element);
        }
    }

    fireEvent(element, eventType, eventProps = {}) {
        const event = new Event(eventType, { bubbles: true, ...eventProps });
        element.dispatchEvent(event);
        return event;
    }

    async waitForElement(selector, timeout = 5000) {
        const startTime = Date.now();
        
        while (Date.now() - startTime < timeout) {
            const element = document.querySelector(selector);
            if (element) return element;
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        
        throw new Error(`Element ${selector} not found within ${timeout}ms`);
    }
}

/**
 * Component Test Utilities
 */
class ComponentTestUtils {
    mount(Component, props = {}, container = null) {
        if (!container) {
            container = document.createElement('div');
            document.body.appendChild(container);
        }

        if (typeof Component === 'function') {
            // Assume it's a class component
            const instance = new Component();
            if (instance.mount) {
                instance.mount(container);
                return { instance, container };
            }
        }

        throw new Error('Unsupported component type');
    }

    unmount(wrapper) {
        if (wrapper.container && wrapper.container.parentNode) {
            wrapper.container.parentNode.removeChild(wrapper.container);
        }
        
        if (wrapper.instance && wrapper.instance.destroy) {
            wrapper.instance.destroy();
        }
    }
}

/**
 * Test Reporter
 */
class TestReporter {
    onStart(testCount) {
        console.log(`\n🧪 Running ${testCount} tests...`);
    }

    onSuiteStart(suiteName) {
        console.log(`\n📁 ${suiteName}`);
    }

    onSuiteEnd(suiteName) {
        // Optional: suite completion message
    }

    onSuiteError(suiteName, error) {
        console.error(`❌ Suite "${suiteName}" failed:`, error);
    }

    onTestStart(description) {
        // Optional: individual test start
    }

    onTestEnd(result) {
        const icon = result.status === 'passed' ? '✅' : '❌';
        const duration = `(${Math.round(result.duration)}ms)`;
        
        console.log(`  ${icon} ${result.description} ${duration}`);
        
        if (result.error) {
            console.error(`    Error: ${result.error.message}`);
        }
    }

    onComplete(results) {
        const summary = this.calculateSummary(results);
        
        console.log(`\n📊 Test Results:`);
        console.log(`   Total: ${summary.total}`);
        console.log(`   Passed: ${summary.passed}`);
        console.log(`   Failed: ${summary.failed}`);
        console.log(`   Success Rate: ${summary.successRate}%`);
        console.log(`   Duration: ${summary.duration}ms`);
        
        if (summary.failed > 0) {
            console.log(`\n❌ ${summary.failed} test(s) failed`);
        } else {
            console.log(`\n🎉 All tests passed!`);
        }
    }

    calculateSummary(results) {
        const total = results.length;
        const passed = results.filter(r => r.status === 'passed').length;
        const failed = results.filter(r => r.status === 'failed').length;
        const duration = results.reduce((sum, r) => sum + r.duration, 0);
        
        return {
            total,
            passed,
            failed,
            successRate: total > 0 ? (passed / total * 100).toFixed(2) : 0,
            duration: Math.round(duration)
        };
    }
}

/**
 * Simple assertion library
 */
class AssertionLibrary {
    constructor(assertions) {
        this.assertions = assertions;
    }

    ok(value, message = 'Expected truthy value') {
        if (!value) {
            throw new Error(message);
        }
        this.assertions.push({ type: 'ok', value, passed: true });
    }

    equal(actual, expected, message = 'Values not equal') {
        if (actual !== expected) {
            throw new Error(`${message}. Expected: ${expected}, Actual: ${actual}`);
        }
        this.assertions.push({ type: 'equal', actual, expected, passed: true });
    }

    throws(fn, message = 'Expected function to throw') {
        try {
            fn();
            throw new Error(message);
        } catch (error) {
            if (error.message === message) {
                throw error; // Re-throw assertion error
            }
            this.assertions.push({ type: 'throws', passed: true });
        }
    }
}

// Create global test framework instance
const testFramework = new TestFramework();

// Export test functions globally
window.describe = testFramework.describe.bind(testFramework);
window.it = testFramework.it.bind(testFramework);
window.testFramework = testFramework;