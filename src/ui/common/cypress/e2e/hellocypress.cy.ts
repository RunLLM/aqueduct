describe('template spec', () => {
  it('passes', () => {
    cy.visit('http://localhost:8080');
    cy.get('input').type('0O54NMDIGET2RJQ1XLVK3YUZB9FC7S86');
    cy.get('button').click()
  })
})