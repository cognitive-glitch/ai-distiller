/**
 * @file Construct 2: The Asynchronous Labyrinth
 * Tests complex asynchronous control flow with Promises, async/await, and Generators
 */

// Generator function
function* idGenerator() {
    let id = 0;
    while(true) {
        yield id++;
    }
}

const gen = idGenerator();

/**
 * Simulates an async data fetch
 * @param {number} id - The ID to fetch
 * @returns {Promise<{id: number, data: string}>} The fetched data
 */
const fetchData = (id) => new Promise(resolve => 
    setTimeout(() => resolve({ id, data: `Data for ${id}` }), 50)
);

/**
 * Main async processing function
 * @returns {Promise<void>}
 */
async function processData() {
    console.log("Starting processing...");
    const firstId = gen.next().value;
    const secondId = gen.next().value;

    // Parallel execution
    const [result1, result2] = await Promise.all([
        fetchData(firstId),
        fetchData(secondId)
    ]);
    console.log("Parallel results:", result1, result2);

    // Sequential execution dependent on previous result
    const thirdId = result1.id + result2.id + gen.next().value;
    try {
        const result3 = await fetchData(thirdId);
        console.log("Sequential result:", result3);
        
        // Nested async operation
        const nestedResult = await (async () => {
            const fourthId = gen.next().value;
            return await fetchData(fourthId);
        })();
        console.log("Nested async result:", nestedResult);
        
    } catch (e) {
        console.error("Error in processing:", e);
    } finally {
        console.log("Processing complete.");
    }
}

// Async generator
async function* asyncDataGenerator() {
    let page = 0;
    while (page < 3) {
        // Simulate fetching paginated data
        await new Promise(resolve => setTimeout(resolve, 100));
        yield { page, items: [`item${page}_1`, `item${page}_2`] };
        page++;
    }
}

/**
 * Consumes async generator
 * @returns {Promise<void>}
 */
async function consumeAsyncData() {
    console.log("Starting async iteration...");
    
    for await (const batch of asyncDataGenerator()) {
        console.log(`Received batch from page ${batch.page}:`, batch.items);
    }
    
    console.log("Async iteration complete.");
}

// Promise race pattern
async function raceExample() {
    const slowPromise = new Promise(resolve => 
        setTimeout(() => resolve('slow'), 200)
    );
    const fastPromise = new Promise(resolve => 
        setTimeout(() => resolve('fast'), 50)
    );
    
    const winner = await Promise.race([slowPromise, fastPromise]);
    console.log(`Winner: ${winner}`);
}

// Error handling with async/await
async function errorHandlingExample() {
    const riskyOperation = () => new Promise((resolve, reject) => {
        if (Math.random() > 0.5) {
            resolve('Success!');
        } else {
            reject(new Error('Random failure'));
        }
    });
    
    try {
        const result = await riskyOperation();
        console.log(result);
    } catch (error) {
        console.error('Caught error:', error.message);
    }
}

// Export all async functions
export {
    processData,
    consumeAsyncData,
    raceExample,
    errorHandlingExample,
    asyncDataGenerator
};