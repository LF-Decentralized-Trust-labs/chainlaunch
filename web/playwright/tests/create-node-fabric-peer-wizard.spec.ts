import { test, expect } from '@playwright/test'
import { login } from './login'

// Helper to generate unique values
function uniqueSuffix() {
  return `${Date.now()}-${Math.floor(Math.random() * 10000)}`
}

const NODE_CREATE_PATH = '/nodes/create'

// This test assumes the admin user is set in env vars for login
// and that at least one organization exists to select

test('can create a Fabric peer node using the NodeCreationWizard', async ({ page, baseURL }) => {
  // Step 1: Login as admin
  await login(page, baseURL ?? '')

  // Step 2: Go to node creation wizard
  await page.goto((baseURL ?? '') + NODE_CREATE_PATH)
  await expect(page.getByRole('heading', { name: /create node/i })).toBeVisible()


  // Step 3: Wizard - Select Protocol (Fabric)
  const nodeName = `test-peer-${uniqueSuffix()}`
  await page.getByPlaceholder('Enter node name').fill(nodeName)
  await page.getByRole('button', { name: 'Fabric' }).click()
  await page.getByRole('button', { name: /next/i }).click()

  // Step 4: Wizard - Select Node Type (Peer)
  await page.getByRole('button', { name: 'Peer node' }).click()
  await page.getByRole('button', { name: /next/i }).click()

  // Step 5: Wizard - Configuration
  await page.getByPlaceholder('Enter node name').fill(nodeName)

  // Select the first available organization (assume dropdown is present)
  const orgSelect = page.getByRole('combobox', { name: /organization/i })
  await orgSelect.click()
  // Select the first option (could be improved to select by name if needed)
  await page.getByRole('option').first().click()

  // Select deployment mode "Docker"
  const modeSelect = page.getByRole('combobox', { name: /mode/i })
  await modeSelect.click()
  await page.getByRole('option', { name: /docker/i }).click()

  // Listen Address
  await page.getByPlaceholder('e.g., 0.0.0.0:7051').fill(`0.0.0.0:${7000 + Math.floor(Math.random() * 1000)}`)
  // Operations Address
  await page.getByPlaceholder('e.g., 0.0.0.0:9443').fill(`0.0.0.0:${9000 + Math.floor(Math.random() * 1000)}`)
  // External Endpoint
  await page.getByPlaceholder('e.g., peer0.org1.example.com:7051').fill(`peer0.example.com:${7000 + Math.floor(Math.random() * 1000)}`)

  // Go to Review step
  await page.getByRole('button', { name: /next/i }).click()

  // Step 6: Review and Submit
  await page.getByRole('button', { name: /create node/i }).click()

  // Wait for navigation to the node detail page or nodes list
  await expect(page.getByText(/General Information/i)).toBeVisible({ timeout: 60000 })
}) 