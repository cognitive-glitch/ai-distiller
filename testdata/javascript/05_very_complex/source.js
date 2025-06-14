/**
 * @file 05_very_complex.js
 * @description Demonstrates highly advanced JavaScript concepts including a custom cooperative
 * task scheduler, advanced generators as coroutines, prototype-based object creation,
 * the Reflect API for metaprogramming, and complex async coordination with Promise.allSettled.
 */

/**
 * A symbol to attach private metadata to a task object.
 * @type {symbol}
 */
const TASK_METADATA = Symbol('taskMetadata');

/**
* Base prototype for all tasks. Using a prototype instead of a class
* allows for more dynamic and flexible object creation and modification at runtime.
* @type {object}
*/
const TaskPrototype = {
   /**
    * Initializes the task.
    * @param {number} id - The unique ID of the task.
    * @param {GeneratorFunction} generatorFunc - The generator function defining the task's work.
    * @param {object} [props={}] - Additional properties for the task.
    */
   init(id, generatorFunc, props = {}) {
       this.id = id;
       this[TASK_METADATA] = {
           status: 'ready',
           createdAt: Date.now(),
           ...props
       };
       // Create the iterator from the generator function. This is the core of the task.
       this.iterator = generatorFunc(this);
       return this;
   },

   /**
    * A reflective method to get a piece of metadata.
    * @param {string} key - The metadata key.
    * @returns {any}
    */
   getMetadata(key) {
       return Reflect.get(this[TASK_METADATA], key);
   },

   /**
    * Private method to update task metadata
    * @private
    * @param {string} key - Metadata key
    * @param {any} value - New value
    */
   _updateMetadata(key, value) {
       Reflect.set(this[TASK_METADATA], key, value);
   }
};

/**
 * @class CooperativeScheduler
 * @description Manages and runs tasks (as generators) in a cooperative, non-blocking manner.
 */
class CooperativeScheduler {
    constructor() {
        /** @private */
        this.taskQueue = [];
        /** @private @type {Map<number, object>} */
        this.tasks = new Map();
        /** @private */
        this.nextTaskId = 1;
        /** @private */
        this.isRunning = false;
        /** @private */
        this._metrics = {
            tasksExecuted: 0,
            totalRuntime: 0,
            averageTaskTime: 0
        };
    }

    /**
     * Registers a new task with the scheduler.
     * @param {GeneratorFunction} generatorFunc - The task's logic.
     * @param {object} [options={}] - Task options
     * @returns {number} The ID of the newly registered task.
     */
    registerTask(generatorFunc, options = {}) {
        const id = this.nextTaskId++;
        // Create a new task object using prototype-based "inheritance".
        const task = Object.create(TaskPrototype).init(id, generatorFunc, options);
        this.tasks.set(id, task);
        this.taskQueue.push(id);
        console.log(`Task ${id} registered.`);
        return id;
    }

    /**
     * Starts the scheduler's event loop.
     */
    start() {
        if (this.isRunning) return;
        this.isRunning = true;
        console.log('Scheduler started.');
        this.run();
    }

    /**
     * The main scheduler loop. This is a complex async method that drives all tasks.
     * @private
     */
    async run() {
        const startTime = performance.now();
        
        while (this.isRunning && this.tasks.size > 0) {
            if (this.taskQueue.length === 0) {
                // Wait for a short period if no tasks are ready to run, preventing a busy loop.
                await new Promise(resolve => setTimeout(resolve, 50));
                continue;
            }

            const taskId = this.taskQueue.shift();
            const task = this.tasks.get(taskId);

            if (!task) continue; // Task might have been deregistered.

            const taskStartTime = performance.now();
            task._updateMetadata('status', 'running');
            task._updateMetadata('lastRun', Date.now());
            
            const result = task.iterator.next();

            if (result.done) {
                // Task is complete, perform cleanup. This is crucial for memory management.
                const taskEndTime = performance.now();
                const taskDuration = taskEndTime - taskStartTime;
                
                console.log(`Task ${task.id} completed in ${taskDuration.toFixed(2)}ms. Cleaning up.`);
                task._updateMetadata('status', 'completed');
                task._updateMetadata('completedAt', Date.now());
                
                this._metrics.tasksExecuted++;
                this.tasks.delete(taskId);
                continue;
            }

            const yieldedValue = result.value;
            if (yieldedValue instanceof Promise) {
                // If the task yielded a promise, wait for it to resolve before requeueing.
                console.log(`Task ${task.id} is awaiting an async operation.`);
                try {
                    await yieldedValue;
                } catch (error) {
                    console.error(`Task ${task.id} encountered an error in async operation:`, error);
                    task.iterator.throw(error); // Propagate error back into the generator
                }
            }
            
            task._updateMetadata('status', 'waiting');
            // Re-queue the task for its next turn (cooperative multitasking).
            this.taskQueue.push(taskId);
        }
        
        const endTime = performance.now();
        this._metrics.totalRuntime = endTime - startTime;
        this._metrics.averageTaskTime = this._metrics.totalRuntime / Math.max(this._metrics.tasksExecuted, 1);
        
        this.isRunning = false;
        console.log('Scheduler stopped. All tasks finished.');
    }

    /**
     * Stops the scheduler gracefully, waiting for all tasks to complete.
     * @param {number} [timeout=5000] - Max time to wait in ms.
     * @returns {Promise<void>}
     */
    async shutdown(timeout = 5000) {
        console.log('Scheduler shutting down gracefully...');
        this.isRunning = false; // Prevent new iterations from starting with new tasks

        const completionPromises = Array.from(this.tasks.values()).map(task => {
            return new Promise(resolve => {
                // This is a simplified shutdown signal. A real implementation might
                // use `task.iterator.return()` to force completion.
                // Here, we just wait for it to be removed from the map.
                const interval = setInterval(() => {
                    if (!this.tasks.has(task.id)) {
                        clearInterval(interval);
                        resolve({ status: 'fulfilled', value: task.id });
                    }
                }, 100);
            });
        });

        const timeoutPromise = new Promise((_, reject) => setTimeout(() => reject(new Error('Shutdown timeout exceeded')), timeout));
        
        const results = await Promise.allSettled([Promise.race([Promise.all(completionPromises), timeoutPromise])]);
        console.log('Shutdown complete.', results);
    }

    /**
     * Gets scheduler performance metrics
     * @returns {object} Performance metrics
     */
    getMetrics() {
        return { ...this._metrics };
    }

    /**
     * Private method to clean up completed tasks
     * @private
     */
    _cleanup() {
        for (const [id, task] of this.tasks) {
            if (task.getMetadata('status') === 'completed') {
                this.tasks.delete(id);
            }
        }
    }

    /**
     * Advanced task introspection using Reflect API
     * @param {number} taskId - Task ID to inspect
     * @returns {object} Task inspection data
     */
    inspectTask(taskId) {
        const task = this.tasks.get(taskId);
        if (!task) return null;

        const properties = Reflect.ownKeys(task);
        const inspection = {
            id: task.id,
            properties: properties.map(prop => ({
                key: prop,
                type: typeof task[prop],
                enumerable: Reflect.getOwnPropertyDescriptor(task, prop)?.enumerable,
                value: prop === TASK_METADATA ? '[PRIVATE METADATA]' : task[prop]
            })),
            metadata: task[TASK_METADATA],
            prototype: Object.getPrototypeOf(task) === TaskPrototype ? 'TaskPrototype' : 'Unknown'
        };

        return inspection;
    }
}

/**
 * Advanced memory pool for object reuse
 * @class MemoryPool
 */
class MemoryPool {
    /**
     * @param {Function} factory - Factory function for creating objects
     * @param {number} maxSize - Maximum pool size
     */
    constructor(factory, maxSize = 10) {
        this._factory = factory;
        this._maxSize = maxSize;
        this._pool = [];
        this._created = 0;
        this._reused = 0;
    }

    /**
     * Acquires object from pool or creates new one
     * @returns {object} Object from pool
     */
    acquire() {
        if (this._pool.length > 0) {
            this._reused++;
            return this._pool.pop();
        }
        
        this._created++;
        return this._factory();
    }

    /**
     * Returns object to pool
     * @param {object} obj - Object to return
     */
    release(obj) {
        if (this._pool.length < this._maxSize) {
            // Reset object state if it has a reset method
            if (typeof obj.reset === 'function') {
                obj.reset();
            }
            this._pool.push(obj);
        }
    }

    /**
     * Gets pool statistics
     * @returns {object} Pool stats
     */
    getStats() {
        return {
            poolSize: this._pool.length,
            created: this._created,
            reused: this._reused,
            efficiency: this._reused / (this._created + this._reused) || 0
        };
    }
}

// --- Usage Example ---

// A helper function that returns a promise, for tasks to yield.
const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Define some tasks as generator functions
function* counterTask() {
    console.log('Counter task started.');
    for (let i = 1; i <= 5; i++) {
        console.log(`Counter: ${i}`);
        yield delay(300); // Yield control and pause for 300ms
    }
    console.log('Counter task finished.');
}

function* dataFetcherTask(task) {
    console.log('Data fetcher task started.');
    console.log(`Fetching data for task ID ${task.id} with metadata:`, Reflect.get(task, TASK_METADATA));
    yield delay(500); // Simulate network request
    console.log('Data fetched.');
    yield delay(200);
    console.log('Processing data...');
    yield; // Yield control for one tick
    console.log('Data fetcher finished.');
}

function* complexIteratorTask() {
    console.log('Complex iterator task started.');
    
    // Custom iterator that yields prime numbers
    const primeIterator = {
        current: 2,
        [Symbol.iterator]() { return this; },
        next() {
            while (!this._isPrime(this.current)) {
                this.current++;
            }
            const value = this.current++;
            return { value, done: value > 20 };
        },
        _isPrime(num) {
            if (num < 2) return false;
            for (let i = 2; i <= Math.sqrt(num); i++) {
                if (num % i === 0) return false;
            }
            return true;
        }
    };

    for (const prime of primeIterator) {
        console.log(`Prime: ${prime}`);
        yield delay(100);
    }
    
    console.log('Complex iterator task finished.');
}

const scheduler = new CooperativeScheduler();
scheduler.registerTask(counterTask);
scheduler.registerTask(dataFetcherTask);
scheduler.registerTask(complexIteratorTask);

scheduler.start();

// After some time, we can register another task dynamically.
setTimeout(() => {
    console.log('\n--- Registering a new task mid-execution ---\n');
    scheduler.registerTask(function* shortLivedTask() {
        console.log('Short-lived task is running!');
        yield;
    });
}, 1000);

// Create memory pool example
const objectPool = new MemoryPool(() => ({ data: null, reset() { this.data = null; } }));

module.exports = { CooperativeScheduler, TaskPrototype, MemoryPool };