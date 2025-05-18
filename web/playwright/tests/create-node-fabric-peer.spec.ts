import { test, expect } from '@playwright/test'
import { login } from './login'
import { createOrganization } from './create-organization'
import { createFabricNode } from './create-node-fabric-peer-logic'

const FABRIC_NODE_CREATE_PATH = '/nodes/fabric/create'

// Helper to generate unique values with cryptographically secure random numbers
function uniqueSuffix() {
	const bytes = new Uint8Array(4);
	crypto.getRandomValues(bytes);
	const randomNum = new DataView(bytes.buffer).getUint32(0) % 10000;
	return `${Date.now()}-${randomNum}`;
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
