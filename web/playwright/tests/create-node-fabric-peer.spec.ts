import { test, expect } from '@playwright/test'
import { login } from './login'
import { createOrganization } from './create-organization'
import { createFabricNode } from './create-node-fabric-peer-logic'

const FABRIC_NODE_CREATE_PATH = '/nodes/fabric/create'

// Helper to generate unique values
function uniqueSuffix() {
	return `${Date.now()}-${Math.floor(Math.random() * 10000)}`
}

test('can login, create an organization, and create a Fabric node', async ({ page, baseURL }) => {
	await login(page, baseURL ?? '')

	// Create a unique organization
	const UNIQUE_SUFFIX = uniqueSuffix()
	const mspId = `test-msp-${UNIQUE_SUFFIX}`
	const description = `Test organization created by Playwright ${UNIQUE_SUFFIX}`
	await createOrganization(page, baseURL, { mspId, description })

	// Use the helper to create a Fabric node
	const nodeName = await createFabricNode(page, baseURL ?? '', mspId)

	// Optionally, assert the node name is visible or other post-creation checks
	await expect(page.getByText(/General Information/i)).toBeVisible({ timeout: 60000 })
})
