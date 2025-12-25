import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Utility function to merge Tailwind CSS classes, handling conflicts.
 * @param inputs - An array of class values.
 * @returns A string of merged class names.
 */
export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };

/**
 * Converts a date or timestamp into a short, locale-specific time string (e.g., "10:30 AM").
 * @param date - The Date object or timestamp to format.
 * @returns The formatted time string.
 */
export function formatShortTime(date: Date | string | number): string {
	const d = new Date(date);
	return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
}

/**
 * Generates initials from a full name.
 * @param name - The full name string.
 * @returns The initials (up to two letters) in uppercase.
 */
export function initials(name: string): string {
	if (!name) return '';
	const parts = name.trim().split(/\s+/);
	if (parts.length === 1) {
		return parts[0][0].toUpperCase();
	}
	return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}
