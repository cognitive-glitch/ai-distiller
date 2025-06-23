// Test Pattern 5: Edge Cases and Complex Patterns
// Tests CommonJS/ES6 mixed imports, prototype modifications, and tricky patterns

// Mixed CommonJS and ES6
const fs = require('fs');
const { promisify } = require('util');
import path from 'path';
import { fileURLToPath } from 'url';

// Modifying prototypes (side-effect that should be kept)
import './polyfills/array-extensions';
import './polyfills/string-extensions';

// Imports with same name from different sources
import { merge } from 'lodash';
import { merge as deepMerge } from 'deepmerge';
import defaultMerge from 'lodash/merge';

// Import that looks like it's unused but is used in template literal
import { version } from './package.json';
import buildNumber from './build-info';

// Import used in eval/Function constructor (hard to detect)
import { validator } from './validators';
import dynamicModule from './dynamic';

// Import used only in comments (should be removed)
import { DocumentationHelper } from './docs';
// See DocumentationHelper for more details

// Not using promisify, path, defaultMerge, buildNumber, dynamicModule, DocumentationHelper

// Using fs (CommonJS)
const data = fs.readFileSync('./data.txt', 'utf8');

// Using fileURLToPath (ES6)
const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Using both merge imports with different names
function mergeConfigs(defaultConfig, userConfig, deep = false) {
  if (deep) {
    return deepMerge(defaultConfig, userConfig);
  }
  return merge(defaultConfig, userConfig);
}

// Using version in template literal
console.log(`App version: ${version}`);

// Using validator in eval (edge case - static analysis might miss this)
const validationCode = 'validator.isEmail("test@example.com")';
const isValid = eval(validationCode);

// Using import in try-catch
try {
  // This makes it harder to detect if the import is used
  const result = new Function('validator', 'return validator.isValid()')(validator);
} catch (error) {
  console.error('Validation failed');
}

// String extensions from polyfill are used (side-effect import)
const formatted = "hello world".capitalize(); // Added by polyfill

export { mergeConfigs };