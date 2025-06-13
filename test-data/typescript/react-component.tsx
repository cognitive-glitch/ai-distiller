// React component with TypeScript

import React, { useState, useEffect, FC } from 'react';
import type { ReactNode, MouseEvent } from 'react';

// Props interfaces
interface ButtonProps {
    children: ReactNode;
    onClick?: (event: MouseEvent<HTMLButtonElement>) => void;
    variant?: 'primary' | 'secondary' | 'danger';
    disabled?: boolean;
}

interface UserCardProps {
    user: {
        id: string;
        name: string;
        email: string;
        avatar?: string;
    };
    onEdit?: (userId: string) => void;
    className?: string;
}

// Generic component props
interface ListProps<T> {
    items: T[];
    renderItem: (item: T, index: number) => ReactNode;
    keyExtractor?: (item: T) => string;
}

// Function component with type
const Button: FC<ButtonProps> = ({ 
    children, 
    onClick, 
    variant = 'primary',
    disabled = false 
}) => {
    const handleClick = (e: MouseEvent<HTMLButtonElement>) => {
        if (!disabled && onClick) {
            onClick(e);
        }
    };
    
    return (
        <button 
            className={`btn btn-${variant}`}
            onClick={handleClick}
            disabled={disabled}
        >
            {children}
        </button>
    );
};

// Class component
class UserCard extends React.Component<UserCardProps> {
    static defaultProps = {
        className: 'user-card'
    };
    
    private handleEdit = () => {
        const { user, onEdit } = this.props;
        if (onEdit) {
            onEdit(user.id);
        }
    };
    
    render() {
        const { user, className } = this.props;
        
        return (
            <div className={className}>
                <h3>{user.name}</h3>
                <p>{user.email}</p>
                <Button onClick={this.handleEdit}>Edit</Button>
            </div>
        );
    }
}

// Generic component
function List<T>({ items, renderItem, keyExtractor }: ListProps<T>) {
    return (
        <ul>
            {items.map((item, index) => (
                <li key={keyExtractor ? keyExtractor(item) : index}>
                    {renderItem(item, index)}
                </li>
            ))}
        </ul>
    );
}

// Custom hooks
function useCounter(initialValue: number = 0): [number, () => void, () => void] {
    const [count, setCount] = useState(initialValue);
    
    const increment = () => setCount(prev => prev + 1);
    const decrement = () => setCount(prev => prev - 1);
    
    return [count, increment, decrement];
}

// HOC with TypeScript
function withAuth<P extends object>(
    Component: React.ComponentType<P>
): React.ComponentType<P & { isAuthenticated: boolean }> {
    return (props) => {
        const isAuthenticated = true; // Simplified
        return <Component {...props} isAuthenticated={isAuthenticated} />;
    };
}

export { Button, UserCard, List, useCounter, withAuth };
export type { ButtonProps, UserCardProps };