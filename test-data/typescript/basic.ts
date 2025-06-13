// Basic TypeScript structures test

// Imports
import React from 'react';
import { Component, useState } from 'react';
import type { User, UserProfile } from './types';
import * as utils from './utils';

// Type aliases
type ID = string | number;
type Callback<T> = (data: T) => void;
type PartialUser = Partial<User>;

// Interfaces
interface IUser {
    id: ID;
    name: string;
    email?: string;
    readonly createdAt: Date;
}

interface IUserService extends IService {
    getUser(id: ID): Promise<IUser>;
    createUser(data: Partial<IUser>): Promise<IUser>;
}

// Enums
enum Status {
    Active = 'ACTIVE',
    Inactive = 'INACTIVE',
    Pending = 'PENDING'
}

enum Priority {
    Low,
    Medium,
    High
}

// Classes
abstract class BaseEntity {
    abstract id: ID;
    abstract validate(): boolean;
    
    protected createdAt: Date = new Date();
    
    constructor(protected readonly type: string) {}
}

class User extends BaseEntity implements IUser {
    id: ID;
    name: string;
    email?: string;
    readonly createdAt: Date;
    
    private _status: Status = Status.Active;
    
    constructor(id: ID, name: string, email?: string) {
        super('user');
        this.id = id;
        this.name = name;
        this.email = email;
        this.createdAt = new Date();
    }
    
    validate(): boolean {
        return this.name.length > 0;
    }
    
    get status(): Status {
        return this._status;
    }
    
    set status(value: Status) {
        this._status = value;
    }
    
    static fromJSON(json: any): User {
        return new User(json.id, json.name, json.email);
    }
}

// Functions with generics
function identity<T>(arg: T): T {
    return arg;
}

async function fetchUser<T extends IUser>(id: ID): Promise<T> {
    const response = await fetch(`/api/users/${id}`);
    return response.json();
}

// Namespace
namespace Utils {
    export function formatDate(date: Date): string {
        return date.toISOString();
    }
    
    export interface IFormatter {
        format(value: any): string;
    }
}

// Type guards
function isUser(obj: any): obj is User {
    return obj instanceof User;
}

// Conditional types
type IsArray<T> = T extends any[] ? true : false;
type ExtractArrayType<T> = T extends (infer U)[] ? U : never;

// Mapped types
type Readonly<T> = {
    readonly [P in keyof T]: T[P];
};

// Export statements
export { User, Status, IUser };
export type { ID, Callback };
export default User;