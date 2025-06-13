// Complex TypeScript test file for AI Distiller functional testing

import { Observable } from 'rxjs';
import { map, filter } from 'rxjs/operators';
import * as fs from 'fs';

// Interfaces
interface User {
    readonly id: number;
    name: string;
    email?: string;
    roles: Role[];
}

interface Repository<T> {
    findById(id: number): Promise<T | null>;
    save(entity: T): Promise<T>;
    delete(id: number): Promise<void>;
}

// Type aliases
type EventHandler<T> = (event: T) => void;
type UserRole = 'admin' | 'user' | 'guest';

// Enums
enum Status {
    Active = 'active',
    Inactive = 'inactive',
    Pending = 'pending'
}

// Classes with inheritance and generics
export abstract class BaseService<T> {
    protected readonly logger: Logger;
    
    constructor(logger: Logger) {
        this.logger = logger;
    }
    
    protected abstract validate(entity: T): boolean;
    
    public async process(entity: T): Promise<T> {
        if (!this.validate(entity)) {
            throw new Error('Validation failed');
        }
        return this.doProcess(entity);
    }
    
    private async doProcess(entity: T): Promise<T> {
        this.logger.info('Processing entity');
        return entity;
    }
}

export class UserService extends BaseService<User> implements Repository<User> {
    private readonly users: Map<number, User> = new Map();
    
    constructor(logger: Logger, private readonly emailService: EmailService) {
        super(logger);
    }
    
    protected validate(user: User): boolean {
        return user.name.length > 0 && user.id > 0;
    }
    
    async findById(id: number): Promise<User | null> {
        return this.users.get(id) || null;
    }
    
    async save(user: User): Promise<User> {
        this.users.set(user.id, user);
        await this.emailService.sendWelcome(user.email || '');
        return user;
    }
    
    async delete(id: number): Promise<void> {
        this.users.delete(id);
    }
    
    // Generic method
    public mapUsers<R>(mapper: (user: User) => R): R[] {
        return Array.from(this.users.values()).map(mapper);
    }
    
    // Static method
    static createDefault(): UserService {
        return new UserService(new ConsoleLogger(), new MockEmailService());
    }
}

// Function with generics and advanced types
export function createObservable<T>(
    source: T[],
    filter?: (item: T) => boolean
): Observable<T> {
    return new Observable<T>(subscriber => {
        const items = filter ? source.filter(filter) : source;
        items.forEach(item => subscriber.next(item));
        subscriber.complete();
    });
}

// Async function
export async function processUserData(
    users: User[],
    options: { validate?: boolean; transform?: boolean } = {}
): Promise<ProcessedUser[]> {
    const { validate = true, transform = false } = options;
    
    return users
        .filter(user => !validate || isValidUser(user))
        .map(user => transform ? transformUser(user) : user as ProcessedUser);
}

// Arrow functions
const isValidUser = (user: User): boolean => {
    return user.name.trim().length > 0 && user.id > 0;
};

const transformUser = (user: User): ProcessedUser => ({
    ...user,
    displayName: `${user.name} (${user.email})`,
    timestamp: Date.now()
});

// Decorators
function Log(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value;
    descriptor.value = function(...args: any[]) {
        console.log(`Calling ${propertyKey} with:`, args);
        return originalMethod.apply(this, args);
    };
}

// Classes with decorators
export class DecoratedService {
    @Log
    public performAction(action: string): void {
        console.log(`Performing action: ${action}`);
    }
    
    private internalMethod(): void {
        // Private implementation
    }
}

// Namespace
namespace Utils {
    export function formatName(first: string, last: string): string {
        return `${first} ${last}`;
    }
    
    export class StringHelper {
        static capitalize(str: string): string {
            return str.charAt(0).toUpperCase() + str.slice(1);
        }
    }
}

// Module augmentation
declare module 'express' {
    interface Request {
        user?: User;
    }
}

// Types and interfaces for complex scenarios
interface ProcessedUser extends User {
    displayName: string;
    timestamp: number;
}

interface Role {
    name: string;
    permissions: Permission[];
}

interface Permission {
    resource: string;
    actions: string[];
}

interface Logger {
    info(message: string): void;
    error(message: string, error?: Error): void;
}

interface EmailService {
    sendWelcome(email: string): Promise<void>;
}

class ConsoleLogger implements Logger {
    info(message: string): void {
        console.log(`[INFO] ${message}`);
    }
    
    error(message: string, error?: Error): void {
        console.error(`[ERROR] ${message}`, error);
    }
}

class MockEmailService implements EmailService {
    async sendWelcome(email: string): Promise<void> {
        console.log(`Sending welcome email to ${email}`);
    }
}

// Export everything for testing
export {
    User, Repository, EventHandler, UserRole, Status,
    Utils, Logger, EmailService, ConsoleLogger, MockEmailService,
    ProcessedUser, Role, Permission
};