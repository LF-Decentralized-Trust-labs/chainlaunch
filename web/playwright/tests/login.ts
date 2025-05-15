import { Page, expect } from '@playwright/test'

const USERNAME = process.env.PLAYWRIGHT_USER
const PASSWORD = process.env.PLAYWRIGHT_PASSWORD
const LOGIN_PATH = '/login'

// Reusable login function
export async function login(page: Page, baseURL: string) {
	await page.goto(baseURL + LOGIN_PATH)
	await expect(page.getByPlaceholder('Enter your username')).toBeVisible()
	await expect(page.getByPlaceholder('Enter your password')).toBeVisible()

	await page.getByPlaceholder('Enter your username').fill(USERNAME || '')
	await page.getByPlaceholder('Enter your password').fill(PASSWORD || '')
	const signInButton = page.getByRole('button', { name: /sign in/i })
	await signInButton.waitFor({ state: 'visible' })
	await signInButton.click()

	// Wait for redirect to any page that is not /login
	// await expect(page).not.toHaveURL(baseURL + LOGIN_PATH, { timeout: 10000 })
	await expect(page).toHaveURL(/.*\/nodes$/, { timeout: 10000 })

	await expect(page.getByRole('heading', { name: 'Nodes' })).toBeVisible({ timeout: 10000 })
	// Optionally, check for a known element on the dashboard or authenticated page
	// await expect(page.getByText('Organizations')).toBeVisible({ timeout: 10000 })
}
