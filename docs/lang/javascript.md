# JavaScript Language Support

AI Distiller provides comprehensive support for JavaScript (ES2015+) and JSX, extracting the essential structure of your code while maintaining type information from JSDoc comments.

## Main Constructs

JavaScript is a dynamically-typed language with several key constructs:

- **Functions** - Regular functions, arrow functions, async functions, generators
- **Classes** - ES6 classes with constructors, methods, and static members
- **Objects** - Object literals with properties and methods
- **Modules** - ES6 imports/exports and CommonJS module.exports
- **Variables** - const, let, var declarations

## Primary Goal of Text Format

The default text format (without stripping) aims to provide AI systems with a complete understanding of available functions, classes, and their signatures. When using `--strip non-public,comments,implementation`, the output focuses on:

- Public API surface (exported functions, classes, methods)
- Function and method signatures with parameter names
- Object structure showing available properties and methods
- Module imports and exports

This gives AI systems enough context to understand what functionality is available and how to use it, without overwhelming them with implementation details.

## Important Features

### Type Information from JSDoc

While JavaScript is dynamically typed, AI Distiller extracts type information from JSDoc comments:

```javascript
/**
 * @param {string} name - User name
 * @param {number} age - User age
 * @returns {User} The created user
 */
function createUser(name, age) { ... }
```

This appears in the distilled output with type annotations, making it easier for AI to understand the expected types.

### Modern JavaScript Support

The parser supports modern JavaScript features including:
- ES6+ syntax (arrow functions, destructuring, template literals)
- Async/await and Promises
- Generators and iterators
- ES6 modules and dynamic imports
- JSX syntax for React components

### Object Literal Analysis

Object literals are parsed to show their structure with full method signatures:
```javascript
const api = {
  name: 'MyAPI',
  getData(id) { ... },
  async processData(input, options) { ... },
  get status() { ... },
  set status(value) { ... }
}
```

Distills to: `+final api = { name, getData(id), processData(input, options), get status(), set status(value) }`

## Known Issues

### Limited Recursive Parsing

The parser uses a shallow parsing approach for performance. This means:
- Nested objects are shown as opaque
- IIFE (Immediately Invoked Function Expressions) are not analyzed for their return values
- Complex destructuring in parameters may be simplified

### CommonJS Support

While `module.exports` is detected and shown, the parser primarily focuses on ES6 modules. Complex CommonJS patterns may not be fully analyzed.

### Dynamic Patterns

JavaScript's dynamic nature means some patterns cannot be statically analyzed:
- Dynamic property access: `obj[variable]`
- Runtime module loading: `require(moduleName)`
- eval() and Function constructor usage

## Real Examples

<details open><summary>Basic Function and Class</summary><blockquote>
  <details><summary>basic.js - source code</summary><blockquote>
    
```javascript
/**
 * @param {string} message - The greeting message
 * @returns {string} Formatted greeting
 */
function greet(message) {
    return `Hello, ${message}!`;
}

class User {
    constructor(name, email) {
        this.name = name;
        this.email = email;
        this._id = Math.random();
    }
    
    getName() {
        return this.name;
    }
    
    async sendEmail(subject, body) {
        console.log(`Sending email to ${this.email}`);
        // Implementation here
    }
}

module.exports = { greet, User };
```
    
  </blockquote></details>
  <details open><summary>Default compact AI-friendly version (`--strip 'non-public,comments,implementation'`)</summary><blockquote>
    
```
<file path="basic.js">
+greet(message: string) -> string

class User
    +constructor(name, email)
    +getName()
    +async sendEmail(subject, body)

# module.exports = { greet, User }
</file>
```
    
  </blockquote></details>
  <details><summary>Full version (`--strip ''`)</summary><blockquote>
    
```
<file path="basic.js">
+greet(message: string) -> string:
    {
        return `Hello, ${message}!`;
    }

class User
    +constructor(name, email):
        {
            this.name = name;
            this.email = email;
            this._id = Math.random();
        }
    
    +getName():
        {
            return this.name;
        }
    
    +async sendEmail(subject, body):
        {
            console.log(`Sending email to ${this.email}`);
            // Implementation here
        }

# module.exports = { greet, User }
</file>
```
    
  </blockquote></details>
</blockquote></details>

<details><summary>React Component with Hooks</summary><blockquote>
  <details><summary>UserList.jsx - source code</summary><blockquote>
    
```javascript
import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';

/**
 * Custom hook for fetching users
 * @returns {Array} List of users
 */
const useUsers = () => {
    const [users, setUsers] = useState([]);
    
    useEffect(() => {
        fetch('/api/users')
            .then(res => res.json())
            .then(data => setUsers(data));
    }, []);
    
    return users;
};

/**
 * User list component
 * @param {Object} props - Component props
 * @param {string} props.title - List title
 * @param {Function} props.onUserClick - Click handler
 */
function UserList({ title, onUserClick }) {
    const users = useUsers();
    
    return (
        <div className="user-list">
            <h2>{title}</h2>
            <ul>
                {users.map(user => (
                    <li key={user.id} onClick={() => onUserClick(user)}>
                        {user.name}
                    </li>
                ))}
            </ul>
        </div>
    );
}

UserList.propTypes = {
    title: PropTypes.string.isRequired,
    onUserClick: PropTypes.func
};

export default UserList;
```
    
  </blockquote></details>
  <details open><summary>Default compact AI-friendly version (`--strip 'non-public,comments,implementation'`)</summary><blockquote>
    
```
<file path="UserList.jsx">
import react
import prop-types

+final useUsers() -> Array

+UserList({ title, onUserClick })

# Exports: UserList
</file>
```
    
  </blockquote></details>
</blockquote></details>

<details><summary>Complex Module with Mixed Patterns</summary><blockquote>
  <details><summary>api-client.js - source code</summary><blockquote>
    
```javascript
import axios from 'axios';

const API_BASE = 'https://api.example.com';

// Private helper
function buildUrl(endpoint) {
    return `${API_BASE}${endpoint}`;
}

/**
 * API client with authentication
 */
class ApiClient {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.axios = axios.create({
            headers: { 'X-API-Key': apiKey }
        });
    }
    
    async get(endpoint) {
        const response = await this.axios.get(buildUrl(endpoint));
        return response.data;
    }
    
    async post(endpoint, data) {
        const response = await this.axios.post(buildUrl(endpoint), data);
        return response.data;
    }
}

// Factory function
export const createClient = (apiKey) => new ApiClient(apiKey);

// Convenience methods
export const quickGet = async (endpoint) => {
    const client = new ApiClient(process.env.API_KEY);
    return client.get(endpoint);
};

export default ApiClient;
```
    
  </blockquote></details>
  <details open><summary>Default compact AI-friendly version (`--strip 'non-public,comments,implementation'`)</summary><blockquote>
    
```
<file path="api-client.js">
import axios

+final API_BASE = 'https://api.example.com'

class ApiClient
    +constructor(apiKey)
    +async get(endpoint)
    +async post(endpoint, data)

+final createClient(apiKey)
+async quickGet(endpoint)

# Exports: createClient, quickGet, ApiClient
</file>
```
    
  </blockquote></details>
</blockquote></details>

## Usage Tips

1. **Use JSDoc comments** for type information - this greatly improves the AI's understanding of your code
2. **Export your public API** explicitly - use named exports or module.exports
3. **Keep object structures simple** - deeply nested objects may not be fully analyzed
4. **Prefer ES6 modules** over CommonJS for better analysis
5. **Use meaningful names** - since implementations can be stripped, good naming is crucial

## Command Examples

```bash
# Analyze a single JavaScript file
aid app.js

# Analyze a React project, showing only public APIs
aid src/ --strip non-public,implementation

# Generate JSON output for tooling integration
aid src/ --format json --output structure.json

# Focus on TypeScript files only
aid src/ --include "*.ts" --exclude "*.test.ts"
```