/**
 * @file Construct 3: The Module System Bridge
 * Tests ESM module with CJS interoperability
 */

import path from 'path'; // Node built-in import
import { fileURLToPath } from 'url';

// ES module specific features
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Modern ESM processor
 * @param {any} data - Data to process
 * @returns {{processed: any, sourceFile: string}} Processed result
 */
export const esmProcessor = (data) => {
    const file = path.basename(import.meta.url);
    return { 
        processed: data, 
        sourceFile: file,
        moduleType: 'ESM'
    };
};

/**
 * Dynamic import of CommonJS module from ESM
 * @returns {Promise<any>} Result from CJS module
 */
export async function dynamicCjsImport() {
    // Dynamic import of a CJS module from an ESM context
    const cjsModule = await import('./cjs_module.js');
    const { legacyProcessor, version, LegacyClass } = cjsModule.default || cjsModule;
    
    console.log(`Dynamically loaded CJS module version: ${version}`);
    
    // Use the imported CJS functionality
    const instance = new LegacyClass('ESM-created');
    return {
        processed: legacyProcessor("dynamic_data"),
        instance: instance.getName(),
        staticCall: LegacyClass.staticMethod()
    };
}

// Named exports
export const MODULE_VERSION = '2.0-esm';

/**
 * @typedef {Object} Config
 * @property {string} environment
 * @property {boolean} debug
 */

/**
 * Get module configuration
 * @returns {Config} Module configuration
 */
export function getConfig() {
    return {
        environment: process.env.NODE_ENV || 'development',
        debug: process.env.DEBUG === 'true'
    };
}

// Re-export pattern
export { esmProcessor as processor } from './construct_3_modules.js';

// Default export
export default {
    esmProcessor,
    dynamicCjsImport,
    MODULE_VERSION,
    getConfig
};