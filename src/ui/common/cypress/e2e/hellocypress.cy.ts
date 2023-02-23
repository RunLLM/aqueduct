describe('template spec', () => {
  it('passes', () => {
    cy.visit('http://localhost:8080');
    cy.get('input').type('O8H9E273PQJ4FWMA61VSDIB0GZNXLKYT');
    cy.get('button').click();
  });
});
