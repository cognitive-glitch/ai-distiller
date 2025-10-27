export interface User {
    id: number;
    name: string;
    email: string;
}

export class UserService {
    private cache: Map<number, User>;
    
    constructor(private apiUrl: string) {
        this.cache = new Map();
    }
    
    async getUser(id: number): Promise<User | null> {
        if (this.cache.has(id)) {
            return this.cache.get(id)!;
        }
        
        const response = await fetch(`${this.apiUrl}/users/${id}`);
        if (!response.ok) {
            return null;
        }
        
        const user = await response.json() as User;
        this.cache.set(id, user);
        return user;
    }
    
    clearCache(): void {
        this.cache.clear();
    }
}
