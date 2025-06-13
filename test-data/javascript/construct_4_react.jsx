/**
 * @file Construct 4: React/JSX Component Tree
 * Tests JSX parsing and React Hook patterns
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import PropTypes from 'prop-types';

/**
 * Custom hook for window width tracking
 * @returns {number} Current window width
 */
const useWindowWidth = () => {
    const [width, setWidth] = useState(window.innerWidth);
    
    useEffect(() => {
        const handleResize = () => setWidth(window.innerWidth);
        window.addEventListener('resize', handleResize);
        
        // Cleanup function
        return () => window.removeEventListener('resize', handleResize);
    }, []); // Empty dependency array
    
    return width;
};

/**
 * Complex React component with multiple features
 * @param {Object} props - Component props
 * @param {string} props.title - Title to display
 * @param {Array<{id: string, name: string}>} props.items - Items to render
 * @param {Function} props.onItemClick - Click handler
 */
function ComplexComponent({ title, items = [], onItemClick }) {
    const [count, setCount] = useState(0);
    const [selectedId, setSelectedId] = useState(null);
    const width = useWindowWidth();
    
    // Memoized computation
    const expensiveValue = useMemo(() => {
        console.log('Computing expensive value...');
        return items.reduce((sum, item) => sum + item.name.length, 0);
    }, [items]);
    
    // Callback with dependencies
    const handleClick = useCallback((itemId) => {
        setSelectedId(itemId);
        setCount(c => c + 1);
        onItemClick?.(itemId);
    }, [onItemClick]);
    
    // Effect with cleanup
    useEffect(() => {
        if (selectedId) {
            console.log(`Selected item: ${selectedId}`);
        }
    }, [selectedId]);
    
    // Conditional classes
    const containerClass = `container ${width < 600 ? 'narrow' : 'wide'}`;
    
    return (
        <div className={containerClass} data-testid="complex-component">
            <header>
                <h1>{title}</h1>
                <span className="counter">Count: {count}</span>
            </header>
            
            {/* Conditional rendering */}
            {count > 5 && (
                <div className="alert">
                    Count is greater than 5!
                </div>
            )}
            
            {/* List rendering with keys */}
            <ul className="item-list">
                {items.map(item => (
                    <li 
                        key={item.id}
                        className={selectedId === item.id ? 'selected' : ''}
                        onClick={() => handleClick(item.id)}
                    >
                        {item.name}
                        {selectedId === item.id && <span> ✓</span>}
                    </li>
                ))}
            </ul>
            
            {/* Fragment usage */}
            <>
                <p>Total characters in names: {expensiveValue}</p>
                <p>Window width: {width}px</p>
            </>
            
            {/* Self-closing component */}
            <ItemCounter count={count} />
        </div>
    );
}

// PropTypes definition
ComplexComponent.propTypes = {
    title: PropTypes.string.isRequired,
    items: PropTypes.arrayOf(PropTypes.shape({
        id: PropTypes.string.isRequired,
        name: PropTypes.string.isRequired
    })),
    onItemClick: PropTypes.func
};

// Simple functional component
const ItemCounter = ({ count }) => (
    <div className="item-counter">
        Items clicked: {count}
    </div>
);

ItemCounter.propTypes = {
    count: PropTypes.number.isRequired
};

// Higher-order component
const withLogging = (WrappedComponent) => {
    const WithLoggingComponent = (props) => {
        useEffect(() => {
            console.log(`${WrappedComponent.name} mounted`);
            return () => console.log(`${WrappedComponent.name} unmounted`);
        }, []);
        
        return <WrappedComponent {...props} />;
    };
    
    WithLoggingComponent.displayName = `withLogging(${WrappedComponent.name})`;
    return WithLoggingComponent;
};

// Export components
export default ComplexComponent;
export { useWindowWidth, ItemCounter, withLogging };