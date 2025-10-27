/**
 * Generic data table with sorting and filtering.
 * Tests: Generics with constraints, complex types, utility types.
 */
import React, { useState, useMemo } from 'react';

interface Column<T> {
    key: keyof T;
    label: string;
    sortable?: boolean;
    render?: (value: T[keyof T], item: T) => React.ReactNode;
}

interface DataTableProps<T extends Record<string, any>> {
    data: T[];
    columns: Column<T>[];
    onRowClick?: (item: T) => void;
    className?: string;
}

type SortDirection = 'asc' | 'desc' | null;

interface SortState<T> {
    key: keyof T | null;
    direction: SortDirection;
}

export function DataTable<T extends Record<string, any>>({
    data,
    columns,
    onRowClick,
    className
}: DataTableProps<T>): JSX.Element {
    const [sortState, setSortState] = useState<SortState<T>>({
        key: null,
        direction: null
    });
    
    const [filter, setFilter] = useState<string>('');
    
    const sortedData = useMemo(() => {
        if (!sortState.key || !sortState.direction) {
            return data;
        }
        
        return [...data].sort((a, b) => {
            const aVal = a[sortState.key!];
            const bVal = b[sortState.key!];
            
            if (aVal < bVal) return sortState.direction === 'asc' ? -1 : 1;
            if (aVal > bVal) return sortState.direction === 'asc' ? 1 : -1;
            return 0;
        });
    }, [data, sortState]);
    
    const filteredData = useMemo(() => {
        if (!filter) return sortedData;
        
        return sortedData.filter(item =>
            Object.values(item).some(value =>
                String(value).toLowerCase().includes(filter.toLowerCase())
            )
        );
    }, [sortedData, filter]);
    
    const handleSort = (key: keyof T) => {
        setSortState(prev => {
            if (prev.key === key) {
                const direction: SortDirection = 
                    prev.direction === 'asc' ? 'desc' : 
                    prev.direction === 'desc' ? null : 'asc';
                return { key: direction ? key : null, direction };
            }
            return { key, direction: 'asc' };
        });
    };
    
    return (
        <div className={className}>
            <input 
                type="text"
                placeholder="Filter..."
                value={filter}
                onChange={e => setFilter(e.target.value)}
            />
            <table>
                <thead>
                    <tr>
                        {columns.map(col => (
                            <th key={String(col.key)} onClick={() => handleSort(col.key)}>
                                {col.label}
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {filteredData.map((item, idx) => (
                        <tr key={idx} onClick={() => onRowClick?.(item)}>
                            {columns.map(col => (
                                <td key={String(col.key)}>
                                    {col.render 
                                        ? col.render(item[col.key], item)
                                        : String(item[col.key])
                                    }
                                </td>
                            ))}
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}
