/**
 * UserProfile component with hooks and TypeScript.
 * Tests: React patterns, hooks, generics, async state.
 */
import React, { useState, useEffect, useMemo } from 'react';

interface User {
    id: number;
    name: string;
    email: string;
    avatar?: string;
}

interface UserProfileProps {
    userId: number;
    onUpdate?: (user: User) => void;
    className?: string;
}

export const UserProfile: React.FC<UserProfileProps> = ({
    userId,
    onUpdate,
    className
}) => {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchUser = async () => {
            try {
                setLoading(true);
                const response = await fetch(`/api/users/${userId}`);
                if (!response.ok) {
                    throw new Error('Failed to fetch user');
                }
                const data = await response.json();
                setUser(data);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
            } finally {
                setLoading(false);
            }
        };

        fetchUser();
    }, [userId]);

    const displayName = useMemo(() => {
        return user ? `${user.name} (${user.email})` : 'Unknown';
    }, [user]);

    if (loading) return <div>Loading...</div>;
    if (error) return <div>Error: {error}</div>;
    if (!user) return <div>User not found</div>;

    return (
        <div className={className}>
            <h2>{displayName}</h2>
            {user.avatar && <img src={user.avatar} alt={user.name} />}
        </div>
    );
};

export default UserProfile;
