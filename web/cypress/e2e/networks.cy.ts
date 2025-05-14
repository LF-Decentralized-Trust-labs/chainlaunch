describe('Network Management', () => {
  beforeEach(() => {
    cy.visit('/networks')
  })

  it('should create a new network', () => {
    // Click create network button
    cy.get('[data-testid="create-network-button"]').click()

    // Fill in network details
    cy.get('[data-testid="network-name-input"]').type('Test Network')
    cy.get('[data-testid="network-type-select"]').select('fabric')
    cy.get('[data-testid="network-description-input"]').type('Test network for e2e testing')

    // Submit the form
    cy.get('[data-testid="create-network-submit"]').click()

    // Verify network was created
    cy.get('[data-testid="network-list"]')
      .should('contain', 'Test Network')
      .and('contain', 'fabric')

    // Verify network status
    cy.get('[data-testid="network-status"]')
      .should('contain', 'ready')
  })

  it('should delete a network', () => {
    // Create a network first
    cy.get('[data-testid="create-network-button"]').click()
    cy.get('[data-testid="network-name-input"]').type('Network to Delete')
    cy.get('[data-testid="network-type-select"]').select('fabric')
    cy.get('[data-testid="network-description-input"]').type('Network that will be deleted')
    cy.get('[data-testid="create-network-submit"]').click()

    // Wait for network to be created
    cy.get('[data-testid="network-list"]').should('contain', 'Network to Delete')

    // Delete the network
    cy.get('[data-testid="network-actions"]').first().click()
    cy.get('[data-testid="delete-network-button"]').click()
    cy.get('[data-testid="confirm-delete-button"]').click()

    // Verify network was deleted
    cy.get('[data-testid="network-list"]').should('not.contain', 'Network to Delete')
  })
}) 