// Test Pattern 1: Basic Imports (Named, Default, Side-Effect)
// Tests standard named, default, and side-effect imports

import { useState, useEffect } from 'react';
import axios from 'axios';
import './styles.css'; // Side-effect import (sets up styles)
import 'core-js/stable'; // Polyfill side-effect import
import { SomeUtil, AnotherUtil } from './utils';
import Logger from './logger';

// Not using useEffect, AnotherUtil, or Logger

function MyComponent() {
  const [count, setCount] = useState(0);
  const [data, setData] = useState(null);

  // Using axios
  const fetchData = async () => {
    try {
      const response = await axios.get('/api/data');
      setData(response.data);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  // Using SomeUtil
  const processedData = SomeUtil.process(data);

  return {
    count,
    setCount,
    fetchData,
    processedData
  };
}