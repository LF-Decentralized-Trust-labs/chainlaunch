import { test, expect } from '@playwright/test'
import { login } from './login'

const FABRIC_NODE_CREATE_PATH = '/nodes/fabric/create'

test('can login and create a Fabric node', async ({ page, baseURL }) => {
	await login(page, baseURL ?? '')

	// Go to Fabric node creation page
	await page.goto(baseURL + FABRIC_NODE_CREATE_PATH)
	await expect(page.getByRole('heading', { name: /create fabric node/i })).toBeVisible({ timeout: 10000 })

	// Fill out the form with unique values
	const UNIQUE_SUFFIX = `${Date.now()}-${Math.floor(Math.random() * 10000)}`
	const nodeName = `test-node-${UNIQUE_SUFFIX}`

	// Name
	await page.getByPlaceholder('Enter node name').fill(nodeName)

	// Select organization (first option)
	const orgSelect = page.getByRole('combobox', { name: /organization/i })
	if (await orgSelect.isVisible().catch(() => false)) {
		await orgSelect.click()
		const firstOrg = page.locator('[role="option"]').first()
		if (await firstOrg.isVisible().catch(() => false)) {
			await firstOrg.click()
		}
	}

	// Node type (default is peer, skip unless you want to test orderer)

	// Listen Address
	await page.getByPlaceholder('e.g., 0.0.0.0:7051').fill(`0.0.0.0:${7000 + Math.floor(Math.random() * 1000)}`)
	// Operations Address
	await page.getByPlaceholder('e.g., 0.0.0.0:9443').fill(`0.0.0.0:${9000 + Math.floor(Math.random() * 1000)}`)
	// External Endpoint
	await page.getByPlaceholder('e.g., peer0.org1.example.com:7051').fill(`peer0.example.com:${7000 + Math.floor(Math.random() * 1000)}`)

	// Submit
	await page.getByRole('button', { name: /create node/i }).click()

	// Wait for navigation to the node detail page or nodes list
	await expect(page).toHaveURL(/\/nodes\//, { timeout: 15000 })
	// Optionally, check for the node name on the detail page
	await expect(page.getByText(/General Information/i)).toBeVisible({ timeout: 10000 })
})
