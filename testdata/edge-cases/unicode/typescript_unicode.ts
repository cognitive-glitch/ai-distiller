/**
 * Edge case: TypeScript with Unicode identifiers and special characters.
 * Tests parser's Unicode handling.
 */

// Unicode in interface names
interface Пользователь {  // Russian: User
    имя: string;  // Russian: name
    возраст: number;  // Russian: age
}

// Emoji in class names (valid in TypeScript/JavaScript)
class 🚀Rocket {
    private 速度: number = 0;  // Chinese: speed

    public 加速(): void {  // Chinese: accelerate
        this.速度++;
    }

    public get現在速度(): number {  // Chinese: current_speed
        return this.速度;
    }
}

// Arabic identifiers
class مستخدم {  // Arabic: User
    constructor(private اسم: string) {}  // Arabic: name

    public الحصول_على_الاسم(): string {  // Arabic: get_name
        return this.اسم;
    }
}

// Greek identifiers
interface Διεπαφή {  // Greek: Interface
    μέθοδος(): void;  // Greek: method
}

// Japanese identifiers
class ユーザー管理 {  // Japanese: User Management
    private ユーザー一覧: string[] = [];  // Japanese: User List

    public ユーザー追加(名前: string): void {  // Japanese: Add User, name
        this.ユーザー一覧.push(名前);
    }
}

// Unicode in type definitions
type 名前型 = string;  // Japanese: Name Type
type 年齢型 = number;  // Japanese: Age Type

// Emoji in function names
function 📊getData(): number[] {
    return [1, 2, 3];
}

function 🔍search(query: string): boolean {
    return query.length > 0;
}

// Unicode in generics
class コンテナ<T> {  // Japanese: Container
    constructor(private 値: T) {}  // Japanese: value

    public 取得(): T {  // Japanese: get
        return this.値;
    }
}

// Zero-width characters
class User​Manager {  // Contains zero-width space (U+200B)
    // Manager with zero-width space
}
