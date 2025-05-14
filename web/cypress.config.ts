import { defineConfig } from 'cypress'

export default defineConfig({
	e2e: {
		baseUrl: process.env.CYPRESS_BASE_URL || 'http://localhost:3100',
		supportFile: 'cypress/support/e2e.ts',
		specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
		video: true,
		screenshotOnRunFailure: true,
		defaultCommandTimeout: 10000,
		viewportWidth: 1280,
		viewportHeight: 720,
		env: {
			apiUrl: process.env.CYPRESS_API_URL || 'http://localhost:8100/api/v1',
		},
	},
})
