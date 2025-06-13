// A type alias for clarity and reusability.
export type UserID = string | number;

/**
 * Checks if a given string is a valid email format.
 * @param email The string to validate.
 * @returns True if the email is valid, false otherwise.
 */
export const isValidEmail = (email: string): boolean => {
  if (!email) return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

// Internal utility function, not exported from the module.
const _normalizeString = (input: string): string => {
  return input.trim().toLowerCase();
};

/**
 * Formats a user ID into a canonical string.
 * @param id The UserID to format.
 * @returns A formatted string.
 */
export const formatUserID = (id: UserID): string => {
  const normalizedId = _normalizeString(String(id));
  return `user_${normalizedId}`;
};