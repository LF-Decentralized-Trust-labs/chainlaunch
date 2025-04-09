import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs))
}

export function slugify(text: string): string {
	return text
		.toString()
		.toLowerCase()
		.replace(/\s+/g, '-')
		.replace(/[^\w\-]+/g, '')
		.replace(/\-\-+/g, '-')
		.replace(/^-+/, '')
		.replace(/-+$/, '')
}

export function numberToHex(num: number): string {
	return `0x${num.toString(16)}`
}

export function hexToNumber(hex: string): number {
	return parseInt(hex.replace('0x', ''), 16)
}

export function isValidHex(hex: string): boolean {
	return /^0x[0-9a-fA-F]+$/.test(hex)
}
