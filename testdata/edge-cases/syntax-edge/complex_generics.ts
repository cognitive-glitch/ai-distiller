/**
 * Edge case: Complex generic constraints and nested generics.
 * Tests parser's handling of advanced TypeScript types.
 */

// Deeply nested generic types
type Nested<T> = T extends Array<infer U>
    ? U extends Array<infer V>
        ? V extends Array<infer W>
            ? W extends Array<infer X>
                ? X
                : W
            : V
        : U
    : T;

// Complex generic constraints
interface Repository<
    T extends { id: number },
    K extends keyof T,
    V extends T[K]
> {
    find(id: number): Promise<T | null>;
    findBy<F extends K>(field: F, value: T[F]): Promise<T[]>;
    save(entity: T): Promise<T>;
}

// Multiple type parameters with constraints
class GenericManager<
    T extends Record<string, any>,
    U extends keyof T,
    V extends T[U],
    W extends Array<V>
> {
    private data: Map<U, W> = new Map();

    public get<K extends U>(key: K): W | undefined {
        return this.data.get(key);
    }

    public set<K extends U>(key: K, value: W): void {
        this.data.set(key, value);
    }
}

// Conditional types with generics
type ExtractPromise<T> = T extends Promise<infer U>
    ? U extends Promise<infer V>
        ? V extends Promise<infer W>
            ? W
            : V
        : U
    : T;

// Mapped types with generics
type DeepReadonly<T> = {
    readonly [P in keyof T]: T[P] extends object
        ? DeepReadonly<T[P]>
        : T[P];
};

// Generic function with multiple constraints
function transform<
    T extends Record<string, any>,
    K extends keyof T,
    R extends Partial<T>
>(
    obj: T,
    key: K,
    transformer: (value: T[K]) => T[K]
): R {
    return {} as R;
}

// Intersection of multiple generic types
type Complex<T, U, V> = (T & U) | (U & V) | (T & V);

// Generic class with static methods
class GenericStatic<T> {
    static create<U>(value: U): GenericStatic<U> {
        return new GenericStatic<U>();
    }

    static transform<U, V>(
        value: U,
        fn: (x: U) => V
    ): V {
        return fn(value);
    }
}
