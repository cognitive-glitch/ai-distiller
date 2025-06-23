// Test Pattern 3: Re-exports, Barrel Exports, and Dynamic Imports
// Tests export...from patterns, barrel exports, and dynamic imports

// Re-exports (these are part of module's public API)
export { default as Button } from './components/Button';
export { Input, TextArea } from './components/Form';
export * from './components/Icons';  // Barrel export
export * as animations from './animations';

// Regular imports
import { helper } from './utils';
import config from './config';
import { Logger } from './logger';
import staticData from './data.json';

// Not using Logger or staticData

// Dynamic imports in functions
async function loadFeature(featureName) {
  // Using config
  if (!config.features[featureName]) {
    return null;
  }
  
  // Dynamic imports based on feature
  switch (featureName) {
    case 'charts':
      const { Chart } = await import('./features/charts');
      return new Chart();
      
    case 'maps':
      const maps = await import('./features/maps');
      return maps.default;
      
    case 'analytics':
      // Dynamic import with error handling
      try {
        const analytics = await import(
          /* webpackChunkName: "analytics" */
          './features/analytics'
        );
        return analytics.initAnalytics();
      } catch (error) {
        console.error('Failed to load analytics:', error);
        return null;
      }
      
    default:
      return null;
  }
}

// Using helper
export function processData(data) {
  return helper.transform(data);
}

// Conditional dynamic import
if (process.env.NODE_ENV === 'development') {
  import('./dev-tools').then(({ setupDevTools }) => {
    setupDevTools();
  });
}