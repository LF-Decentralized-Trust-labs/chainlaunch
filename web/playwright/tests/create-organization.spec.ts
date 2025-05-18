import { test } from '@playwright/test'
import { createOrganization } from './create-organization'
import { login } from './login'

test('can login and create a new organization', async ({ page, baseURL }) => {
	await login(page, baseURL ?? '')

	// Use a unique suffix with cryptographically secure random values
	const bytes = new Uint8Array(4)
	crypto.getRandomValues(bytes)
	const randomNum = new DataView(bytes.buffer).getUint32(0) % 10000
	const UNIQUE_SUFFIX = `${Date.now()}-${randomNum}`
	const mspId = `test-msp-${UNIQUE_SUFFIX}`
	const description = `Test organization created by Playwright ${UNIQUE_SUFFIX}`
	await createOrganization(page, baseURL, { mspId, description })
})
