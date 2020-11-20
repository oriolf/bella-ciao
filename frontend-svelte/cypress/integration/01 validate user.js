describe('Admin user', () => {
	beforeEach(() => {
        cy.visit('/');
        cy.get('#unique_id').type("11111111H");
        cy.get('#password').type("12345678");
        cy.get('button').contains('Entra').click();
	});

	it('shows the list of existing users', () => {
		cy.contains('h6', '11111111H');
		cy.contains('h6', '22222222J');
		cy.contains('h6', '33333333P');
		cy.contains('h6', 'X1111111G');
    });
    
    // TODO validate previously registered user
});
