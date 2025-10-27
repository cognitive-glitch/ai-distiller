// Test Pattern 2: Aliased Imports, Namespace Imports, and Type-Only Imports
// Tests import aliases, namespace imports, and TypeScript type imports

import * as React from 'react';
import { Component as ReactComponent } from 'react';
import { func as myFunc, helper as myHelper } from './helpers';
import type { UserProfile, Settings } from './types';
import { type Config, ConfigManager } from './config';
import * as Utils from './utils';
import lodash from 'lodash';
import { map, filter, reduce } from 'lodash';

// Type import used in generics
import type { Validator } from './validators';

// Not using ReactComponent, myHelper, Settings, Utils namespace, filter, reduce

interface Props {
  user: UserProfile;  // Using type-only import
  config: Config;     // Using mixed type/value import
}

function validateUser<T extends Validator<UserProfile>>(
  user: UserProfile,
  validator: T
): boolean {
  return validator.validate(user);
}

class App extends React.Component<Props> {
  render() {
    // Using React namespace import
    return React.createElement('div', null, 'Hello');
  }

  processData() {
    // Using myFunc alias
    const result = myFunc([1, 2, 3]);

    // Using lodash default import and named import map
    const numbers = [1, 2, 3, 4, 5];
    const doubled = map(numbers, n => n * 2);
    const sum = lodash.sum(doubled);

    // Using ConfigManager (value import from mixed import)
    const manager = new ConfigManager();

    return { result, sum, manager };
  }
}