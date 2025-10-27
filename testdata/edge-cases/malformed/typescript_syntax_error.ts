/**
 * Edge case: TypeScript file with syntax errors.
 * Parser should handle gracefully without crashing.
 */

// Valid interface
interface ValidUser {
    id: number;
    name: string;
}

// Syntax error: Missing closing brace
function brokenFunction(x: number, y: number): number {
    return x + y;
// Missing }

// Syntax error: Unclosed generic
class GenericClass<T extends Record {
    value: T;
}

// Valid class after errors
class ValidClass {
    private field: string = "test";

    public method(): void {
        console.log("still parsing");
    }
}

// Syntax error: Invalid type annotation
function badTypes(x:: number): void {
    console.log(x);
}

// Syntax error: Missing semicolon and malformed arrow function
const arrow = (x: number) => {
    const y = x + 1
    const z = y + 2 =>  // Invalid arrow here
}
